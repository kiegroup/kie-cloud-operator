import React from "react";
import { FormGroup, TextInput } from "@patternfly/react-core";

export class DefaultTextField {
  constructor(props) {
    this.props = props;
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
          onChange={this.onChangeText}
          jsonpath={this.props.fieldDef.jsonPath}
          defaultValue={this.props.fieldDef.value}
          {...this.props.attrs}
        />
      </FormGroup>
    );
  }
  onChangeText = value => {
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
    this.props.fieldDef.errMsg = this.errMsg;
  }
}
