import React, { Component } from "react";

import { Wizard } from "@patternfly/react-core";

export default class OperatorWizard extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isOpen: true,
      allStepsValid: true
    };
    this.onGoToStep = this.onGoToStep.bind(this);
    this.areAllStepsValid = this.areAllStepsValid.bind(this);
    this.onNext = this.onNext.bind(this);
    this.onBack = this.onBack.bind(this);
    this.onGoToStep = this.onGoToStep.bind(this);
  }

  /*
   * TODO: these events should address the state change of the Wizard.
   * After each change, the state must be persisted in the local storage and
   * bring back as soon as the user navigates to the Page
   */

  areAllStepsValid = () => {
    this.setState({
      allStepsValid: false
    });
  };

  onNext = ({ id, name, component }, { prevId, prevName }) => {
    console.log(
      `current id: ${id}, current name: ${name}, previous id: ${prevId}, previous name: ${prevName}`
    );
    console.log(` component: ${component}`);
    this.areAllStepsValid();
  };

  onBack = ({ id, name }, { prevId, prevName }) => {
    console.log(
      `current id: ${id}, current name: ${name}, previous id: ${prevId}, previous name: ${prevName}`
    );
    this.areAllStepsValid();
  };

  onGoToStep = ({ id, name }, { prevId, prevName }) => {
    console.log(
      `current id: ${id}, current name: ${name}, previous id: ${prevId}, previous name: ${prevName}`
    );
  };

  render() {
    return (
      <Wizard
        isOpen={true}
        title="Operator GUI"
        description="KIE Operator"
        isFullHeight
        isFullWidth
        onClose={() => {}}
        steps={this.props.steps}
        onNext={this.onNext}
        onBack={this.onBack}
        onGoToStep={this.onGoToStep}
      />
    );
  }
}
