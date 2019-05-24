import React, { Component } from "react";

import { Wizard, TextArea, Button, Modal, Alert } from "@patternfly/react-core";
import YAML from "js-yaml";
import Dot from "dot-object";
import CopyToClipboard from "react-copy-to-clipboard";

import OperatorWizardFooter from "./OperatorWizardFooter";
import { BACKEND_URL } from "../common/GuiConstants";
import { loadJsonSpec } from "./FormJsonLoader";
import StepBuilder from "./StepBuilder";

export default class OperatorWizard extends Component {
  constructor(props) {
    super(props);
    this.title = "Operator installer";
    this.subtitle = "RHPAM installer";
    this.stepBuilder = new StepBuilder();
    this.state = {
      isOpen: true,
      steps: this.stepBuilder.buildPlaceholderStep(),
      isFormValid: false,
      currentStep: 1,
      maxSteps: 1,
      isModalOpen: false,
      isDeployModalOpen: false,
      isErrorModalOpen: false
    };
    document.title = this.title;

    loadJsonSpec().then(spec =>
      this.setState({
        spec: spec
      })
    );

    this.stepBuilder.buildSteps().then(result => {
      this.setState({
        steps: result.steps,
        pages: result.pages,
        maxSteps: this.calculateSteps(result.steps)
      });
    });
  }

  calculateSteps = pages => {
    let steps = 0;
    pages.forEach(p => {
      if (p.steps !== undefined) {
        steps += this.calculateSteps(p.steps);
      } else {
        steps++;
      }
    });
    return steps;
  };

  onPageChange = ({ id }) => {
    this.setState({
      currentStep: id
    });
  };

  onDeploy = () => {
    if (!this.validateForm()) {
      console.log("The form has validation errors");
      this.handleErrorModalToggle();
      return;
    }
    const result = this.createResultYaml();
    this.handleDeployModalToggle();
    console.log(result);
    fetch(BACKEND_URL, {
      method: "POST",
      body: JSON.stringify(result),
      headers: {
        "Content-Type": "application/yaml"
      }
    })
      .then(res => res.json())
      .then(response => console.log("Success:", JSON.stringify(response)))
      .catch(error => console.error("Error:", error));
  };

  onEditYaml = () => {
    if (!this.validateForm()) {
      console.log("The form has validation errors");
      this.handleErrorModalToggle();
      return;
    }
    this.createResultYaml();
    this.handleModalToggle();
  };

  handleModalToggle = () => {
    this.setState(({ isModalOpen }) => ({
      isModalOpen: !isModalOpen
    }));
  };

  handleErrorModalToggle = () => {
    this.setState(({ isErrorModalOpen }) => ({
      isErrorModalOpen: !isErrorModalOpen
    }));
  };

  handleDeployModalToggle = () => {
    this.setState(({ isDeployModalOpen }) => ({
      isDeployModalOpen: !isDeployModalOpen
    }));
  };

  onChangeYaml = resultYaml => {
    this.setState({
      resultYaml
    });
  };

  //TODO: Validation should only be done onFieldUpdated and only for this field
  validateForm() {
    let isValid = true;
    this.state.pages.forEach(page => {
      if (page.subPages !== undefined) {
        page.subPages.forEach(subPage => {
          if (!this.validateFields(subPage.fields)) {
            isValid = false;
          }
        });
      }
      if (isValid) {
        isValid = this.validateFields(page.fields);
      }
    });
    this.setState({
      isFormValid: isValid
    });
    return isValid;
  }

  validateFields(fields) {
    let isValid = true;
    if (fields !== undefined) {
      fields.forEach(field => {
        if (
          (field.type === "object" ||
            ((field.type === "dropDown" || field.type === "fieldGroup") &&
              field.fields !== undefined &&
              (field.visible !== undefined && field.visible !== false))) &&
          !this.validateFields(field.fields)
        ) {
          isValid = false;
          return;
        }
        if (field.errMsg !== undefined && field.errMsg !== "") {
          console.log(`Field ${field.label} is not valid: ${field.errMsg}`);
          isValid = false;
        }
      });
    }
    return isValid;
  }

  createYamlFromPages() {
    let jsonObject = {};

    if (Array.isArray(this.state.pages)) {
      this.state.pages.forEach(page => {
        let pageFields = page.fields;

        if (Array.isArray(pageFields)) {
          pageFields.forEach(field => {
            if (
              field.type === "dropDown" &&
              field.fields !== undefined &&
              field.visible !== false
            ) {
              jsonObject = this.addObjectFields(field, jsonObject);
            }
            if (field.type === "object" || field.type === "fieldGroup") {
              jsonObject = this.addObjectFields(field, jsonObject);
            } else {
              const value =
                field.type === "checkbox" ? field.checked : field.value;
              if (
                field.jsonPath !== undefined &&
                field.jsonPath !== "" &&
                value !== undefined &&
                value !== ""
              ) {
                let jsonPath = this.getJsonSchemaPathForYaml(field.jsonPath);
                jsonObject[jsonPath] = value;
              }
            }
          });
        }
        if (
          page.subPages !== undefined &&
          Array.isArray(page.subPages) &&
          page.subPages.length > 0
        ) {
          page.subPages.forEach(subPage => {
            let subPageFields = subPage.fields;

            subPageFields.forEach(field => {
              if (
                field.type === "dropDown" &&
                field.fields !== undefined &&
                field.visible !== false
              ) {
                jsonObject = this.addObjectFields(field, jsonObject);
              }
              if (field.type === "object" || field.type === "fieldGroup") {
                jsonObject = this.addObjectFields(field, jsonObject);
              } else {
                const value =
                  field.type === "checkbox" ? field.checked : field.value;
                if (
                  field.jsonPath !== undefined &&
                  field.jsonPath !== "" &&
                  value !== undefined &&
                  value !== ""
                ) {
                  let jsonPath = this.getJsonSchemaPathForYaml(field.jsonPath);
                  jsonObject[jsonPath] = value;
                }
              }
            });
          });
        }
      });
    }
    return jsonObject;
  }

  addObjectFields(field, jsonObject) {
    if (Array.isArray(field.fields)) {
      field.fields.forEach(field => {
        if (
          field.type === "dropDown" &&
          field.fields !== undefined &&
          field.visible !== false
        ) {
          jsonObject = this.addObjectFields(field, jsonObject);
        }
        if (field.type === "object" || field.type === "fieldGroup") {
          jsonObject = this.addObjectFields(field, jsonObject);
        } else {
          const value = field.type === "checkbox" ? field.checked : field.value;
          if (
            field.jsonPath !== undefined &&
            field.jsonPath !== "" &&
            value !== undefined &&
            value !== "" &&
            field.visible !== false
          ) {
            let jsonPath = this.getJsonSchemaPathForYaml(field.jsonPath);
            jsonObject[jsonPath] = value;
          }
        }
      });
    }
    return jsonObject;
  }

  getJsonSchemaPathForYaml(jsonPath) {
    return jsonPath.slice(2, jsonPath.length);
  }

  createResultYaml() {
    const jsonObject = this.createYamlFromPages();
    var resultYaml =
      "apiVersion: " +
      this.state.spec.apiVersion +
      "\n" +
      "kind: " +
      this.state.spec.kind +
      "\n";
    if (Object.getOwnPropertyNames(jsonObject).length > 0) {
      resultYaml = resultYaml + YAML.safeDump(Dot.object(jsonObject));
    }
    this.setState({
      resultYaml: resultYaml
    });
    return resultYaml;
  }

  render() {
    const operatorFooter = (
      <OperatorWizardFooter
        // isFormValid={this.state.isFormValid}
        isFormValid={true} //TODO: Replace by line above when validation works properly
        maxSteps={this.state.maxSteps}
        onDeploy={this.onDeploy}
        onEditYaml={this.onEditYaml}
        onNext={this.onPageChange}
        onBack={this.onPageChange}
        onGoToStep={this.onPageChange}
      />
    );
    return (
      <React.Fragment>
        <Wizard
          isOpen={true}
          title={this.title}
          description={this.subtitle}
          isFullHeight
          isFullWidth
          onClose={() => {}}
          steps={this.state.steps}
          footer={operatorFooter}
        />
        <Modal
          title=" "
          isOpen={this.state.isModalOpen}
          onClose={this.handleModalToggle}
          actions={[
            <CopyToClipboard
              key="yaml_copy"
              className="pf-c-button pf-m-primary"
              onCopy={this.onCopyYaml}
              text={this.state.resultYaml}
            >
              <button key="yaml_button_copy">Copy to clipboard</button>
            </CopyToClipboard>,
            <Button
              key="cancel"
              variant="secondary"
              onClick={this.handleModalToggle}
            >
              Cancel
            </Button>
          ]}
        >
          <TextArea
            id="yaml_edit_text"
            key="yaml_text"
            onChange={this.onChangeYaml}
            rows={100}
            cols={35}
            value={this.state.resultYaml}
          />
        </Modal>

        <Modal
          isSmall
          title=""
          isOpen={this.state.isDeployModalOpen}
          onClose={this.handleDeployModalToggle}
        >
          <Alert variant="info" title="Deployment request was created." />
        </Modal>

        <Modal
          isSmall
          title=""
          isOpen={this.state.isErrorModalOpen}
          onClose={this.handleErrorModalToggle}
        >
          <Alert variant="danger" title="Validation errors!" />
        </Modal>
      </React.Fragment>
    );
  }
}
