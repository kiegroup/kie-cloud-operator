import React, { Component } from "react";
import validator from "validator";
import { FormGroup, TextInput } from "@patternfly/react-core";

export class EmailField extends Component {
  constructor(props) {
    super(props);

    this.state = {
      value: this.props.fieldDef.value,
      isValid: true,
      errMsg: this.props.fieldDef.errMsg
    };
    this.props = props;

    this.handleTextInputChange = value => {
      this.isValidField(value);
      this.props.fieldDef.value = value;
      this.props.fieldDef.errMsg = this.state.errMsg;
    };
  }

  getJsx() {
    let { value } = this.state;

    let { isValid, errMsg } = this.validate(value);

    return (
      <FormGroup
        label={this.props.fieldDef.label}
        fieldId={this.props.ids.fieldGroupId}
        key={this.props.ids.fieldGroupKey}
        helperTextInvalid={errMsg}
        helperText={this.props.fieldDef.description}
        isValid={isValid}
        isRequired={this.props.fieldDef.required}
      >
        <TextInput
          type="text"
          id={this.props.ids.fieldId}
          key={this.props.ids.fieldKey}
          aria-describedby="horizontal-form-name-helper"
          name={this.props.fieldDef.label}
          onChange={this.handleTextInputChange}
          jsonpath={this.props.fieldDef.jsonPath}
          defaultValue={value}
          {...this.props.attrs}
        />
      </FormGroup>
    );
  }
  isValidField(value) {
    let { isValid, errMsg } = this.validate(value);
    this.setState({ value, isValid, errMsg });
  }
  validate(value) {
    let isValid = true;
    let errMsg = "";
    if (this.props.fieldDef.required === true && value === "") {
      errMsg = this.props.fieldDef.label + " is required.";
      isValid = false;
    } else if (
      value !== undefined &&
      value !== "" &&
      !validator.isEmail(value)
    ) {
      errMsg = value + " is not a valid email.";
      isValid = false;
    } else {
      errMsg = "";
      isValid = true;
    }
    this.props.fieldDef.errMsg = errMsg;
    return { isValid, errMsg };
  }

  render() {
    return this.getJsx();
  }
}
