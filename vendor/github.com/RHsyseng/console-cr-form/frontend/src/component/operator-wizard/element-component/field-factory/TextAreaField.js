import React from "react";
import { FormGroup, TextArea } from "@patternfly/react-core";

export class TextAreaField {
  constructor(props) {
    this.props = props;
    this.onBlurTextArea = this.onBlurTextArea.bind(this);
    this.errMsg = "";
    this.isValid = true;
  }

  getJsx() {
    this.isValidField();
    return (
      <FormGroup
        label={this.props.fieldDef.label}
        fieldId={this.props.ids.fieldGroupId}
        key={this.props.ids.fieldGroupKey}
        helperTextInvalid={this.errMsg}
        helperText={this.props.fieldDef.description}
        isValid={this.isValid}
        isRequired={this.props.fieldDef.required}
      >
        <TextArea
          value={this.props.fieldDef.default}
          name="horizontal-form-exp"
          id={this.props.ids.fieldId}
          key={this.props.ids.fieldKey}
          jsonpath={this.props.fieldDef.jsonPath}
          defaultValue={this.props.fieldDef.value}
          onBlur={this.onBlurTextArea}
        />
      </FormGroup>
    );
  }

  onBlurTextArea = event => {
    let value = event.target.value;
    if (value !== undefined && value !== null) {
      this.props.fieldDef.value = value;
      this.isValidField();
    }
  };

  isValidField() {
    const value = this.props.fieldDef.value;
    if (
      this.props.fieldDef.required === true &&
      (value === undefined || value === "")
    ) {
      this.errMsg = this.props.fieldDef.label + " is required.";
      this.isValid = false;
    } else {
      this.errMsg = "";
      this.isValid = true;
    }
  }
}
