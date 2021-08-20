package connection

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stackrox/rox/central/networkpolicies/graph"
	"github.com/stackrox/rox/central/scrape"
	"github.com/stackrox/rox/central/sensor/networkentities"
	"github.com/stackrox/rox/central/sensor/networkpolicies"
	"github.com/stackrox/rox/central/sensor/service/common"
	"github.com/stackrox/rox/central/sensor/service/pipeline"
	"github.com/stackrox/rox/central/sensor/telemetry"
	"github.com/stackrox/rox/generated/internalapi/central"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/booleanpolicy/policyversion"
	"github.com/stackrox/rox/pkg/centralsensor"
	"github.com/stackrox/rox/pkg/concurrency"
	"github.com/stackrox/rox/pkg/logging"
	"github.com/stackrox/rox/pkg/reflectutils"
	"github.com/stackrox/rox/pkg/sac"
	"github.com/stackrox/rox/pkg/stringutils"
	"github.com/stackrox/rox/pkg/sync"
)

var (
	log = logging.LoggerForModule()
)

type sensorConnection struct {
	clusterID           string
	stopSig, stoppedSig concurrency.ErrorSignal

	sendC chan *central.MsgToSensor

	scrapeCtrl          scrape.Controller
	networkPoliciesCtrl networkpolicies.Controller
	networkEntitiesCtrl networkentities.Controller
	telemetryCtrl       telemetry.Controller

	sensorEventHandler *sensorEventHandler

	queues      map[string]*dedupingQueue
	queuesMutex sync.Mutex

	eventPipeline pipeline.ClusterPipeline

	clusterMgr         common.ClusterManager
	networkEntityMgr   common.NetworkEntityManager
	policyMgr          common.PolicyManager
	baselineMgr        common.ProcessBaselineManager
	networkBaselineMgr common.NetworkBaselineManager

	sensorHello  *central.SensorHello
	capabilities centralsensor.SensorCapabilitySet
}

func newConnection(sensorHello *central.SensorHello,
	cluster *storage.Cluster,
	eventPipeline pipeline.ClusterPipeline,
	clusterMgr common.ClusterManager,
	networkEntityMgr common.NetworkEntityManager,
	policyMgr common.PolicyManager,
	baselineMgr common.ProcessBaselineManager,
	networkBaselineMgr common.NetworkBaselineManager,
) *sensorConnection {

	conn := &sensorConnection{
		stopSig:       concurrency.NewErrorSignal(),
		stoppedSig:    concurrency.NewErrorSignal(),
		sendC:         make(chan *central.MsgToSensor),
		eventPipeline: eventPipeline,
		queues:        make(map[string]*dedupingQueue),

		clusterID:          cluster.GetId(),
		clusterMgr:         clusterMgr,
		policyMgr:          policyMgr,
		networkEntityMgr:   networkEntityMgr,
		baselineMgr:        baselineMgr,
		networkBaselineMgr: networkBaselineMgr,

		sensorHello:  sensorHello,
		capabilities: centralsensor.CapSetFromStringSlice(sensorHello.GetCapabilities()...),
	}

	// Need a reference to conn for injector
	conn.sensorEventHandler = newSensorEventHandler(eventPipeline, conn, &conn.stopSig)
	conn.scrapeCtrl = scrape.NewController(conn, &conn.stopSig)
	conn.networkPoliciesCtrl = networkpolicies.NewController(conn, &conn.stopSig)
	conn.networkEntitiesCtrl = networkentities.NewController(cluster.GetId(), networkEntityMgr, graph.Singleton(), conn, &conn.stopSig)
	conn.telemetryCtrl = telemetry.NewController(conn.capabilities, conn, &conn.stopSig)

	return conn
}

func (c *sensorConnection) Terminate(err error) bool {
	return c.stopSig.SignalWithError(err)
}

func (c *sensorConnection) Stopped() concurrency.ReadOnlyErrorSignal {
	return &c.stoppedSig
}

// multiplexedPush pushes the given message to a dedicated queue for the respective event type.
// The queues parameter, if non-nil, will be used to look up the queue by event type. If the `queues`
// map is nil or does not contain an entry for the respective type, a queue is retrieved from the
// mutex-protected `c.queues` map (and created if exists), and afterwards stored in the `queues` map
// if non-nil.
// The envisioned use for this is that a caller invoking `multiplexedPush` repeatedly will maintain
// an exclusively used (i.e., not requiring protection via a mutex) map, that will automatically be
// populated with a subset of the entries from `c.queues`. This avoids mutex lock acquisitions for every
// invocation of `multiplexedPush` with a previously seen (from the perspective of the caller)
// event type.
func (c *sensorConnection) multiplexedPush(ctx context.Context, msg *central.MsgFromSensor, queues map[string]*dedupingQueue) {
	typ := reflectutils.Type(msg.Msg)
	queue := queues[typ]
	if queue == nil {
		concurrency.WithLock(&c.queuesMutex, func() {
			queue = c.queues[typ]
			if queue == nil {
				queue = newDedupingQueue(stripTypePrefix(typ))
				go c.handleMessages(ctx, queue)
				c.queues[typ] = queue
			}
		})
		if queues != nil {
			queues[typ] = queue
		}
	}
	queue.push(msg)
}

func (c *sensorConnection) runRecv(ctx context.Context, grpcServer central.SensorService_CommunicateServer) {
	queues := make(map[string]*dedupingQueue)
	for !c.stopSig.IsDone() {
		msg, err := grpcServer.Recv()
		if err != nil {
			c.stopSig.SignalWithError(errors.Wrap(err, "recv error"))
			return
		}

		c.multiplexedPush(ctx, msg, queues)
	}
}

func (c *sensorConnection) handleMessages(ctx context.Context, queue *dedupingQueue) {
	for msg := queue.pullBlocking(&c.stopSig); msg != nil; msg = queue.pullBlocking(&c.stopSig) {
		if err := c.handleMessage(ctx, msg); err != nil {
			log.Errorf("Error handling sensor message: %v", err)
		}
	}
	c.eventPipeline.OnFinish(c.clusterID)
	c.stoppedSig.SignalWithError(c.stopSig.Err())
}

func (c *sensorConnection) runSend(server central.SensorService_CommunicateServer) {
	for !c.stopSig.IsDone() {
		select {
		case <-c.stopSig.Done():
			return
		case <-server.Context().Done():
			c.stopSig.SignalWithError(errors.Wrap(server.Context().Err(), "context error"))
			return
		case msg := <-c.sendC:
			if err := server.Send(msg); err != nil {
				c.stopSig.SignalWithError(errors.Wrap(err, "send error"))
				return
			}
		}
	}
}

func (c *sensorConnection) Scrapes() scrape.Controller {
	return c.scrapeCtrl
}

func (c *sensorConnection) InjectMessageIntoQueue(msg *central.MsgFromSensor) {
	c.multiplexedPush(sac.WithAllAccess(withConnection(context.Background(), c)), msg, nil)
}

func (c *sensorConnection) NetworkEntities() networkentities.Controller {
	return c.networkEntitiesCtrl
}

func (c *sensorConnection) NetworkPolicies() networkpolicies.Controller {
	return c.networkPoliciesCtrl
}

func (c *sensorConnection) Telemetry() telemetry.Controller {
	return c.telemetryCtrl
}

func (c *sensorConnection) InjectMessage(ctx concurrency.Waitable, msg *central.MsgToSensor) error {
	select {
	case c.sendC <- msg:
		return nil
	case <-ctx.Done():
		return errors.New("context aborted")
	case <-c.stopSig.Done():
		return errors.Wrap(c.stopSig.Err(), "could not send message as sensor connection was stopped")
	}
}

func (c *sensorConnection) handleMessage(ctx context.Context, msg *central.MsgFromSensor) error {
	switch m := msg.Msg.(type) {
	case *central.MsgFromSensor_ScrapeUpdate:
		return c.scrapeCtrl.ProcessScrapeUpdate(m.ScrapeUpdate)
	case *central.MsgFromSensor_NetworkPoliciesResponse:
		return c.networkPoliciesCtrl.ProcessNetworkPoliciesResponse(m.NetworkPoliciesResponse)
	case *central.MsgFromSensor_TelemetryDataResponse:
		return c.telemetryCtrl.ProcessTelemetryDataResponse(m.TelemetryDataResponse)
	case *central.MsgFromSensor_Event:
		// Special case the reprocess deployment because its fields are already set
		if msg.GetEvent().GetReprocessDeployment() != nil {
			c.sensorEventHandler.addMultiplexed(ctx, msg)
			return nil
		}
		// Only dedupe on non-creates
		if msg.GetEvent().GetAction() != central.ResourceAction_CREATE_RESOURCE {
			msg.DedupeKey = msg.GetEvent().GetId()
		}
		// Set the hash key for all values
		msg.HashKey = msg.GetEvent().GetId()

		c.sensorEventHandler.addMultiplexed(ctx, msg)
		return nil
	}
	return c.eventPipeline.Run(ctx, msg, c)
}

// getPolicySyncMsg fetches stored policies and prepares them for delivery to sensor.
func (c *sensorConnection) getPolicySyncMsg(ctx context.Context) (*central.MsgToSensor, error) {
	policies, err := c.policyMgr.GetAllPolicies(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting policies for initial sync")
	}

	return c.getPolicySyncMsgFromPolicies(policies)
}

// getPolicySyncMsgFromPolicies prepares given policies for delivery to sensor. If:
//   - sensor's policy version is unknown -> guess Version1,
//   - sensor's policy version is older than central's -> attempt to downgrade
//     all policies to sensor's version,
//   - otherwise -> forward policies unmodified.
//
// If there is any error during the downgrade, pass the affected policies as-is.
func (c *sensorConnection) getPolicySyncMsgFromPolicies(policies []*storage.Policy) (*central.MsgToSensor, error) {
	// Older sensors do not broadcast the policy version they support, so if we
	// observe an empty string, we guess the version at Version1 and persist it.
	sensorPolicyVersionStr := stringutils.FirstNonEmpty(c.sensorHello.GetPolicyVersion(), policyversion.Version1().String())

	// Forward policies as is if we don't understand sensor's version.
	if sensorPolicyVersion, err := policyversion.FromString(sensorPolicyVersionStr); err != nil {
		log.Errorf("Cannot understand sensor's policy version %q: %v", sensorPolicyVersionStr, err)
	} else {
		// Downgrade all policies if necessary. If we can't downgrade one,
		// we likely can't convert any of them, so no need to spam the log.
		if policyversion.Compare(policyversion.CurrentVersion(), sensorPolicyVersion) > 0 {
			log.Infof("Downgrading %d policies from central's version %q to sensor's version %q",
				len(policies), policyversion.CurrentVersion().String(), sensorPolicyVersion.String())

			downgradedPolicies := make([]*storage.Policy, 0, len(policies))

			var downgradeErr error
			for _, p := range policies {
				cloned := p.Clone()
				downgradedPolicies = append(downgradedPolicies, cloned)
				err := policyversion.DowngradePolicyTo(cloned, sensorPolicyVersion)
				if downgradeErr == nil && err != nil {
					downgradeErr = err
				}
			}
			if downgradeErr != nil {
				log.Errorf("Policy downgrade failed: %v", downgradeErr)
			}

			policies = downgradedPolicies
		}
	}

	return &central.MsgToSensor{
		Msg: &central.MsgToSensor_PolicySync{
			PolicySync: &central.PolicySync{
				Policies: policies,
			},
		},
	}, nil
}

func (c *sensorConnection) getNetworkBaselineSyncMsg(ctx context.Context) (*central.MsgToSensor, error) {
	var networkBaselines []*storage.NetworkBaseline
	err := c.networkBaselineMgr.Walk(ctx, func(baseline *storage.NetworkBaseline) error {
		if !baseline.GetLocked() {
			// Baseline not locked yet. No need to sync to sensor
			return nil
		}
		if baseline.GetClusterId() != c.clusterID {
			// Not a baseline of the cluster we are talking to
			return nil
		}
		networkBaselines = append(networkBaselines, baseline)
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not list network baselines for Sensor connection")
	}
	return &central.MsgToSensor{
		Msg: &central.MsgToSensor_NetworkBaselineSync{
			NetworkBaselineSync: &central.NetworkBaselineSync{
				NetworkBaselines: networkBaselines,
			},
		},
	}, nil
}

func (c *sensorConnection) getBaselineSyncMsg(ctx context.Context) (*central.MsgToSensor, error) {
	var baselines []*storage.ProcessBaseline
	err := c.baselineMgr.WalkAll(ctx, func(pw *storage.ProcessBaseline) error {
		if pw.GetUserLockedTimestamp() == nil {
			return nil
		}
		if pw.GetKey().GetClusterId() != c.clusterID {
			return nil
		}
		baselines = append(baselines, pw)
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not list process baselines for Sensor connection")
	}
	return &central.MsgToSensor{
		Msg: &central.MsgToSensor_BaselineSync{
			BaselineSync: &central.BaselineSync{
				Baselines: baselines,
			},
		},
	}, nil
}

func (c *sensorConnection) getClusterConfigMsg(ctx context.Context) (*central.MsgToSensor, error) {
	cluster, exists, err := c.clusterMgr.GetCluster(ctx, c.clusterID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.Errorf("could not pull config for cluster %q because it does not exist", c.clusterID)
	}
	return &central.MsgToSensor{
		Msg: &central.MsgToSensor_ClusterConfig{
			ClusterConfig: &central.ClusterConfig{
				Config: cluster.GetDynamicConfig(),
			},
		},
	}, nil
}

func (c *sensorConnection) getAuditLogSyncMsg(ctx context.Context) (*central.MsgToSensor, error) {
	cluster, exists, err := c.clusterMgr.GetCluster(ctx, c.clusterID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.Errorf("could not pull config for cluster %q because it does not exist", c.clusterID)
	}

	return &central.MsgToSensor{
		Msg: &central.MsgToSensor_AuditLogSync{
			AuditLogSync: &central.AuditLogSync{
				NodeAuditLogFileStates: cluster.GetAuditLogState(),
			},
		},
	}, nil
}

func (c *sensorConnection) Run(ctx context.Context, server central.SensorService_CommunicateServer, connectionCapabilities centralsensor.SensorCapabilitySet) error {
	// Synchronously send the config to ensure syncing before Sensor marks the connection as Central reachable
	msg, err := c.getClusterConfigMsg(ctx)
	if err != nil {
		return errors.Wrapf(err, "unable to get cluster config for %q", c.clusterID)
	}

	if err := server.Send(msg); err != nil {
		return errors.Wrapf(err, "unable to sync config to cluster %q", c.clusterID)
	}

	if connectionCapabilities.Contains(centralsensor.SensorDetectionCap) {
		msg, err = c.getPolicySyncMsg(ctx)
		if err != nil {
			return errors.Wrapf(err, "unable to get policy sync msg for %q", c.clusterID)
		}
		if err := server.Send(msg); err != nil {
			return errors.Wrapf(err, "unable to sync initial policies to cluster %q", c.clusterID)
		}

		msg, err = c.getBaselineSyncMsg(ctx)
		if err != nil {
			return errors.Wrapf(err, "unable to get process baseline sync msg for %q", c.clusterID)
		}
		if err := server.Send(msg); err != nil {
			return errors.Wrapf(err, "unable to sync initial process baselines to cluster %q", c.clusterID)
		}

		msg, err = c.getNetworkBaselineSyncMsg(ctx)
		if err != nil {
			return errors.Wrapf(err, "unable to get network baseline sync msg for %q", c.clusterID)
		}
		if err := server.Send(msg); err != nil {
			return errors.Wrapf(err, "unable to sync initial network baselines to cluster %q", c.clusterID)
		}

	}

	go c.runSend(server)

	// Trigger initial network graph external sources sync. Network graph external sources capability is added to sensor only if the the feature is enabled.
	if connectionCapabilities.Contains(centralsensor.NetworkGraphExternalSrcsCap) {
		if err := c.NetworkEntities().SyncNow(ctx); err != nil {
			log.Errorf("Unable to sync initial external network entities to cluster %q: %v", c.clusterID, err)
		}
	}

	if connectionCapabilities.Contains(centralsensor.AuditLogEventsCap) {
		msg, err := c.getAuditLogSyncMsg(ctx)
		if err != nil {
			return errors.Wrapf(err, "unable to get audit log file state sync msg for %q", c.clusterID)
		}

		// Send the audit log state to sensor even if the the user has it disabled (that's set in dynamic config). When enabled, sensor will use it correctly
		if err := server.Send(msg); err != nil {
			return errors.Wrapf(err, "unable to sync audit log file state to cluster %q", c.clusterID)
		}
	}

	c.runRecv(ctx, server)
	return c.stopSig.Err()
}

func (c *sensorConnection) ClusterID() string {
	return c.clusterID
}

func (c *sensorConnection) HasCapability(capability centralsensor.SensorCapability) bool {
	return c.capabilities.Contains(capability)
}

func (c *sensorConnection) ObjectsDeletedByReconciliation() (map[string]int, bool) {
	return c.sensorEventHandler.reconciliationMap.DeletedElementsByType()
}

func (c *sensorConnection) CheckAutoUpgradeSupport() error {
	if c.sensorHello.GetHelmManagedConfigInit() != nil && !c.sensorHello.GetHelmManagedConfigInit().GetNotHelmManaged() {
		return errors.New("cluster is Helm-managed and does not support auto upgrades; use 'helm upgrade' or a Helm-aware CD pipeline for upgrades")
	}
	return nil
}

func (c *sensorConnection) SensorVersion() string {
	return c.sensorHello.GetSensorVersion()
}
