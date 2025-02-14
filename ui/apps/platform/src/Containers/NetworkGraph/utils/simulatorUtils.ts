import { NetworkPolicyModification } from 'Containers/Network/networkTypes';

export function getDisplayYAMLFromNetworkPolicyModification(
    modification: NetworkPolicyModification | null
): string {
    const { applyYaml, toDelete } = modification || {};
    const shouldDelete = toDelete && toDelete.length > 0;
    const showApplyYaml = applyYaml && applyYaml.length >= 2;

    const toDeleteSection = shouldDelete
        ? toDelete
              .map((entry) => `# kubectl -n ${entry.namespace} delete networkpolicy ${entry.name}`)
              .join('\n')
        : '';

    // Format complete YAML for display.
    let displayYaml: string;
    if (shouldDelete && showApplyYaml) {
        displayYaml = [toDeleteSection, applyYaml].join('\n---\n');
    } else if (shouldDelete && !showApplyYaml) {
        displayYaml = toDeleteSection;
    } else if (!shouldDelete && showApplyYaml) {
        displayYaml = applyYaml;
    } else {
        displayYaml = 'No policies need to be created or deleted.';
    }

    return displayYaml;
}
