import React from "react";

import { FormGroup, Radio } from "@patternfly/react-core";

export class RadioButtonField {
  constructor(props) {
    this.props = props;
  }

  doGenerateJsx() {
    const fieldIdTrue = this.props.ids.fieldId + "-true";
    const fieldKeyTrue = this.props.ids.fieldKey + "-true";
    const fieldIdFalse = this.props.ids.fieldId + "-false";
    const fieldKeyFalse = this.props.ids.fieldKey + "-false";
    return (
      <FormGroup
        label={this.props.fieldDef.label}
        fieldId={this.props.ids.fieldGroupId}
        key={this.props.ids.fieldGroupKey}
      >
        <Radio
          label="Yes"
          aria-label="radio yes"
          id={fieldIdTrue}
          key={fieldKeyTrue}
          name="horizontal-radios"
          jsonpath={this.props.fieldDef.jsonPath}
        />
        <Radio
          label="No"
          aria-label="radio no"
          id={fieldIdFalse}
          key={fieldKeyFalse}
          name="horizontal-radios"
          jsonpath={this.props.fieldDef.jsonPath}
        />
      </FormGroup>
    );
  }
}
