import React from "react";
import validator from "validator";
import { FormGroup, TextInput } from "@patternfly/react-core";

export class EmailField {
  constructor(props) {
    this.props = props;
    this.onBlurText = this.onBlurText.bind(this);
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
        <TextInput
          type="text"
          id={this.props.ids.fieldId}
          key={this.props.ids.fieldKey}
          aria-describedby="horizontal-form-name-helper"
          name={this.props.fieldDef.label}
          onBlur={this.onBlurText}
          jsonpath={this.props.fieldDef.jsonPath}
          defaultValue={this.props.fieldDef.value}
          {...this.props.attrs}
        />
      </FormGroup>
    );
  }

  onBlurText = event => {
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
    } else if (value !== "" && !validator.isEmail(value)) {
      this.errMsg = value + " is not a valid email.";
      this.isValid = false;
    } else {
      this.errMsg = "";
      this.isValid = true;
    }
  }
}
