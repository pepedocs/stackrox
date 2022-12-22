import {
    Dropdown,
    DropdownDirection,
    DropdownItem,
    DropdownToggle,
    Split,
    SplitItem,
} from '@patternfly/react-core';
import React from 'react';
import UploadYAMLButton from './UploadYAMLButton';

type NetworkSimulatorActionsProps = {
    generateNetworkPolicies: () => void;
    undoNetworkPolicies: () => void;
    onFileInputChange: (
        _event: React.ChangeEvent<HTMLInputElement> | React.DragEvent<HTMLElement>,
        file: File
    ) => void;
    applyNetworkPolicies?: () => void;
};

const actionsDropdownId = 'network-simulator-actions-dropdown';

const labels = {
    generate: 'Rebuild rules from active traffic',
    undo: 'Revert rules to previously applied YAML',
    apply: 'Apply network policies',
    share: 'Share YAML with notifiers',
};

function NetworkSimulatorActions({
    generateNetworkPolicies,
    undoNetworkPolicies,
    onFileInputChange,
    applyNetworkPolicies,
}: NetworkSimulatorActionsProps) {
    const [isActionsOpen, setIsActionsOpen] = React.useState(false);

    const onToggle = (isOpen: boolean) => {
        setIsActionsOpen(isOpen);
    };

    const onFocus = () => {
        const element = document.getElementById(actionsDropdownId);
        element?.focus();
    };

    const onSelect = () => {
        setIsActionsOpen(false);
        onFocus();
    };

    const actionsDropdownItems = [
        <DropdownItem key="generate" tooltip="" onClick={generateNetworkPolicies}>
            {labels.generate}
        </DropdownItem>,
        <DropdownItem key="undo" tooltip="" onClick={undoNetworkPolicies}>
            {labels.undo}
        </DropdownItem>,
    ];

    if (applyNetworkPolicies) {
        actionsDropdownItems.unshift(
            <DropdownItem key="apply" tooltip="" onClick={applyNetworkPolicies}>
                {labels.apply}
            </DropdownItem>
        );
    }

    return (
        <Split hasGutter className="pf-u-p-md">
            <SplitItem>
                <UploadYAMLButton onFileInputChange={onFileInputChange} />
            </SplitItem>
            <SplitItem>
                <Dropdown
                    direction={DropdownDirection.up}
                    position="left"
                    onSelect={onSelect}
                    toggle={
                        <DropdownToggle id={actionsDropdownId} onToggle={onToggle}>
                            Actions
                        </DropdownToggle>
                    }
                    isOpen={isActionsOpen}
                    dropdownItems={actionsDropdownItems}
                />
            </SplitItem>
        </Split>
    );
}

export default NetworkSimulatorActions;
