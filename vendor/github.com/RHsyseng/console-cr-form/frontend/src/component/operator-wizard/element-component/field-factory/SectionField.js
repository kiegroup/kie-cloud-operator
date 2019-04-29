import React from "react";

import { FormGroup, Button } from "@patternfly/react-core";
import { PlusCircleIcon, MinusCircleIcon } from "@patternfly/react-icons";
import FieldFactory from "./FieldFactory";

export class SectionField {
  constructor(props) {
    this.props = props;
  }

  getJsx() {
    var section = this.props.fieldDef.label + "section";
    var jsxArray = [];
    var iconIdPlus = section + "plus";
    var iconIdMinus = section + "minus";

    var fieldJsx = (
      <FormGroup
        fieldId={this.props.ids.fieldGroupId}
        key={this.props.ids.fieldGroupKey}
      >
        <Button
          variant="link"
          id={this.props.ids.fieldId}
          key={this.props.ids.fieldKey}
          fieldnumber={this.props.ids.fieldNumber}
          onClick={this.expandSection}
          name={section}
          style={{ display: "inline-block" }}
        >
          {this.props.fieldDef.label}
          <PlusCircleIcon
            key={"plus-" + this.props.ids.fieldKey}
            id={iconIdPlus}
            style={{ display: "block" }}
          />{" "}
          <MinusCircleIcon
            key={"minus-" + this.props.ids.fieldKey}
            id={iconIdMinus}
            style={{ display: "none" }}
          />
        </Button>
      </FormGroup>
    );

    var children = FieldFactory.newInstancesAsJsx(
      this.props.fieldDef.fields,
      null
    );
    jsxArray.push(fieldJsx);
    fieldJsx = (
      <div id={section} key={section} style={{ display: "none" }}>
        {children}
      </div>
    );
    jsxArray.push(fieldJsx);
    return jsxArray;
  }

  expandSection = event => {
    const target = event.target;
    const name = target.name;

    var elem = document.getElementById(name);
    //TODO: figure out a way to not manipulate the DOM directly
    if (elem.style.display === "block") {
      elem.style.display = "none";
      document.getElementById(name + "plus").style.display = "block";
      document.getElementById(name + "minus").style.display = "none";
    } else {
      elem.style.display = "block";
      document.getElementById(name + "plus").style.display = "none";
      document.getElementById(name + "minus").style.display = "block";
    }
  };
}
