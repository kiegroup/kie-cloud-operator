import React from "react";
import { FormGroup, TextInput } from "@patternfly/react-core";

export class DefaultTextField {
  constructor(props) {
    this.props = props;
    this.onBlurText = this.onBlurText.bind(this);
  }

  getJsx() {
    if (this.props.fieldDef.required === true) {
      return (
        <FormGroup
          label={this.props.fieldDef.label}
          fieldId={this.props.ids.fieldGroupId}
          key={this.props.ids.fieldGroupKey}
          isRequired
        >
          <TextInput
            isRequired
            type="text"
            id={this.props.ids.fieldId}
            key={this.props.ids.fieldKey}
            aria-describedby="horizontal-form-name-helper"
            name={this.props.fieldDef.label}
            // onChange={this.onChangeText}
            onBlur={this.onBlurText}
            jsonpath={this.props.fieldDef.jsonPath}
            // value={((this.props.fieldDef.default!==undefined ) ? this.props.fieldDef.default:this.props.fieldDef.value)}
            defaultValue={this.props.fieldDef.value}
            //value={this.value}
            {...this.props.attrs}
          />
        </FormGroup>
      );
    } else {
      return (
        <FormGroup
          label={this.props.fieldDef.label}
          fieldId={this.props.ids.fieldGroupId}
          key={this.props.ids.fieldGroupKey}
        >
          <TextInput
            type="text"
            id={this.props.ids.fieldId}
            key={this.props.ids.fieldKey}
            aria-describedby="horizontal-form-name-helper"
            name={this.props.fieldDef.label}
            // onChange={this.onChangeText}
            onBlur={this.onBlurText}
            jsonpath={this.props.fieldDef.jsonPath}
            // value={((this.props.fieldDef.default!==undefined ) ? this.props.fieldDef.default:this.props.fieldDef.value)}
            defaultValue={this.props.fieldDef.value}
            //value={this.value}
            {...this.props.attrs}
          />
        </FormGroup>
      );
    }
  }

  onBlurText = event => {
    let value = event.target.value;
    if (value !== undefined && value !== null && value !== "") {
      this.props.fieldDef.value = value;
      // this.props.fieldDef.default="";

      //document.getElementById(event.target.id).value = value;
    }
  };
}
