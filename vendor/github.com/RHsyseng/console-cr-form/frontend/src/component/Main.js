import React, { Component } from "react";
import OperatorWizard from "./operator-wizard/OperatorWizard";
import StepBuilder from "./operator-wizard/StepBuilder";

export default class Main extends Component {
  constructor(props) {
    super(props);
    this.stepBuilder = new StepBuilder();
    this.state = {
      steps: [this.stepBuilder.buildPlaceholderStep()],
      showPopup: false
    };
  }

  componentDidMount() {
    this.setState({ steps: this.stepBuilder.buildSteps() });
  }

  render() {
    return (
      <React.Fragment>
        <OperatorWizard steps={this.state.steps} />
      </React.Fragment>
    );
  }
}
