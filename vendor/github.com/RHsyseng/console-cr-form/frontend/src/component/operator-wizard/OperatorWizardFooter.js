import React from "react";
import PropTypes from "prop-types";
import {
  WizardFooter,
  WizardContextConsumer,
  Button
} from "@patternfly/react-core";

export default class OperatorWizardFooter extends React.Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <WizardFooter>
        <WizardContextConsumer>
          {({ activeStep, onNext, onBack, goToStepById }) => {
            const goToReview = () => {
              if (this.props.validate()) {
                goToStepById(this.props.maxSteps);
              } else {
                goToStepById(this.props.getErrorStep());
              }
            };

            const nextBtn = (
              <Button variant="primary" type="submit" onClick={onNext}>
                Next
              </Button>
            );
            const onViewYamlBtn = () => {
              if (this.props.validate()) {
                this.props.onEditYaml();
              } else {
                goToStepById(this.props.getErrorStep());
              }
            };

            const onDeployBtn = () => {
              if (this.props.validate()) {
                this.props.onDeploy();
              } else {
                goToStepById(this.props.getErrorStep());
              }
            };

            const deployBtn = (
              <Button variant="primary" type="submit" onClick={onDeployBtn}>
                Deploy
              </Button>
            );

            const backBtn = (
              <Button
                variant="secondary"
                type="submit"
                onClick={onBack}
                className={activeStep.id === 1 ? "pf-m-disabled" : ""}
              >
                Back
              </Button>
            );

            const viewYamlBtn = (
              <Button
                variant="link"
                isInline
                onClick={onViewYamlBtn}
                // className={this.props.isFormValid ? "" : "pf-m-disabled"}
              >
                View YAML
              </Button>
            );

            const finishBtn = (
              <Button
                variant="secondary"
                type="submit"
                onClick={goToReview}
                // className={this.props.isFormValid ? "" : "pf-m-disabled"}
              >
                Finish
              </Button>
            );

            return (
              <React.Fragment>
                {this.props.isFinished
                  ? ""
                  : activeStep.id !== this.props.maxSteps
                  ? nextBtn
                  : deployBtn}
                {!this.props.isFinished && backBtn}
                {viewYamlBtn}
                {activeStep.id !== this.props.maxSteps ? finishBtn : ""}
              </React.Fragment>
            );
          }}
        </WizardContextConsumer>
      </WizardFooter>
    );
  }
}

OperatorWizardFooter.propTypes = {
  maxSteps: PropTypes.number.isRequired,
  isFormValid: PropTypes.bool.isRequired,
  validate: PropTypes.func.isRequired, // TODO: Remove when validation is
  onDeploy: PropTypes.func.isRequired,
  onEditYaml: PropTypes.func.isRequired,
  isFinished: PropTypes.bool.isRequired
};
