import React, { Component } from "react";

import { Wizard } from "@patternfly/react-core";

export default class OperatorWizard extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isOpen: true,
      isFormValidA: true,
      formValueA: "Five",
      isFormValidB: true,
      formValueB: "Six",
      allStepsValid: true
    };
    this.toggleOpen = this.toggleOpen.bind(this);
    this.onFormChangeA = this.onFormChangeA.bind(this);
    this.onFormChangeB = this.onFormChangeB.bind(this);
    this.onGoToStep = this.onGoToStep.bind(this);
    this.areAllStepsValid = this.areAllStepsValid.bind(this);
    this.onNext = this.onNext.bind(this);
    this.onBack = this.onBack.bind(this);
    this.onGoToStep = this.onGoToStep.bind(this);
    this.onSave = this.onSave.bind(this);
  }

  /*
   * TODO: these events should address the state change of the Wizard.
   * After each change, the state must be persisted in the local storage and
   * bring back as soon as the user navigates to the Page
   */

  toggleOpen = () => {
    this.setState(({ isOpen }) => ({
      isOpen: !isOpen
    }));
  };

  onFormChangeA = (isValid, value) => {
    this.setState(
      {
        isFormValidA: isValid,
        formValueA: value
      },
      this.areAllStepsValid
    );
  };

  onFormChangeB = (isValid, value) => {
    this.setState(
      {
        isFormValidB: isValid,
        formValueB: value
      },
      this.areAllStepsValid
    );
  };

  areAllStepsValid = () => {
    this.setState({
      allStepsValid: this.state.isFormValidA && this.state.isFormValidB
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

  onSave = () => {
    console.log("Saved and closed the wizard");
    this.setState({
      isOpen: false
    });
  };

  render() {
    return (
      <Wizard
        isOpen={true}
        title="Operator GUI"
        description="KIE Operator"
        onClose={this.toggleOpen}
        onSave={this.onSave}
        steps={this.props.steps}
        onNext={this.onNext}
        onBack={this.onBack}
        onGoToStep={this.onGoToStep}
      />
    );
  }
}
