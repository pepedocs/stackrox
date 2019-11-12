import React from 'react';
import gql from 'graphql-tag';

import { defaultHeaderClassName, defaultColumnClassName } from 'Components/Table';
import TopCvssLabel from 'Components/TopCvssLabel';
import WorkflowListPage from 'Containers/Workflow/WorkflowListPage';
import entityTypes from 'constants/entityTypes';
import CVEStackedPill from 'Components/CVEStackedPill';
import TableCountLink from 'Components/workflow/TableCountLink';
import queryService from 'modules/queryService';

import { VULN_COMPONENT_LIST_FRAGMENT } from 'Containers/VulnMgmt/VulnMgmt.fragments';
import { workflowListPropTypes, workflowListDefaultProps } from 'constants/entityPageProps';
import removeEntityContextColumns from 'utils/tableUtils';

export const defaultComponentSort = [
    {
        id: 'priority',
        desc: false
    },
    {
        id: 'name',
        desc: false
    }
];

export function getComponentTableColumns(workflowState) {
    const tableColumns = [
        {
            Header: 'Id',
            headerClassName: 'hidden',
            className: 'hidden',
            accessor: 'id'
        },
        {
            Header: `Component`,
            headerClassName: `w-1/4 ${defaultHeaderClassName}`,
            className: `w-1/4 ${defaultColumnClassName}`,
            Cell: ({ original }) => {
                const { version, name } = original;
                return `${name} ${version}`;
            },
            accessor: 'name'
        },
        {
            Header: `CVEs`,
            entityType: entityTypes.CVE,
            headerClassName: `w-1/8 ${defaultHeaderClassName}`,
            className: `w-1/8 ${defaultColumnClassName}`,
            Cell: ({ original, pdf }) => {
                const { vulnCounter, id } = original;
                if (!vulnCounter || vulnCounter.all.total === 0) return 'No CVEs';

                const newState = workflowState.pushListItem(id).pushList(entityTypes.CVE);
                const url = newState.toUrl();

                // If `Fixed By` is set, it means vulnerability is fixable.
                const fixableUrl = newState.setSearch({ 'Fixed By': 'r/.*' }).toUrl();

                return (
                    <CVEStackedPill
                        vulnCounter={vulnCounter}
                        url={url}
                        fixableUrl={fixableUrl}
                        hideLink={pdf}
                    />
                );
            },
            accessor: 'vulnCounter.all.total'
        },
        {
            Header: `Top CVSS`,
            headerClassName: `w-1/10 text-center ${defaultHeaderClassName}`,
            className: `w-1/10 ${defaultColumnClassName}`,
            Cell: ({ original }) => {
                const { topVuln } = original;
                if (!topVuln)
                    return (
                        <div className="mx-auto flex flex-col">
                            <span>–</span>
                        </div>
                    );
                const { cvss, scoreVersion } = topVuln;
                return <TopCvssLabel cvss={cvss} version={scoreVersion} />;
            },
            accessor: 'topVuln.cvss'
        },
        {
            Header: `Images`,
            entityType: entityTypes.IMAGE,
            headerClassName: `w-1/8 ${defaultHeaderClassName}`,
            className: `w-1/8 ${defaultColumnClassName}`,
            accessor: 'imageCount',
            Cell: ({ original, pdf }) => (
                <TableCountLink
                    entityType={entityTypes.IMAGE}
                    count={original.imageCount}
                    textOnly={pdf}
                    selectedRowId={original.id}
                />
            )
        },
        {
            Header: `Deployments`,
            entityType: entityTypes.DEPLOYMENT,
            headerClassName: `w-1/8 ${defaultHeaderClassName}`,
            className: `w-1/8 ${defaultColumnClassName}`,
            accessor: 'deploymentCount',
            Cell: ({ original, pdf }) => (
                <TableCountLink
                    entityType={entityTypes.DEPLOYMENT}
                    count={original.deploymentCount}
                    textOnly={pdf}
                    selectedRowId={original.id}
                />
            )
        },
        {
            Header: `Risk Priority`,
            headerClassName: `w-1/10 ${defaultHeaderClassName}`,
            className: `w-1/10 ${defaultColumnClassName}`,
            accessor: 'priority'
        }
    ];

    return removeEntityContextColumns(tableColumns, workflowState);
}

const VulnMgmtComponents = ({ selectedRowId, search, sort, page, data }) => {
    const query = gql`
        query getComponents($query: String) {
            results: components(query: $query) {
                ...componentFields
            }
        }
        ${VULN_COMPONENT_LIST_FRAGMENT}
    `;

    const queryOptions = {
        variables: {
            query: queryService.objectToWhereClause(search)
        }
    };

    return (
        <WorkflowListPage
            data={data}
            query={query}
            queryOptions={queryOptions}
            idAttribute="id"
            entityListType={entityTypes.COMPONENT}
            defaultSorted={sort || defaultComponentSort}
            getTableColumns={getComponentTableColumns}
            selectedRowId={selectedRowId}
            page={page}
            search={search}
        />
    );
};

VulnMgmtComponents.propTypes = workflowListPropTypes;
VulnMgmtComponents.defaultProps = {
    ...workflowListDefaultProps,
    sort: null
};

export default VulnMgmtComponents;
