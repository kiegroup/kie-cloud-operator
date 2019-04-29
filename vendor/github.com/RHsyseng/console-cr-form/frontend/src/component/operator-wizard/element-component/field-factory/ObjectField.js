import React from "react";
import { ActionGroup, Button } from "@patternfly/react-core";
import FieldFactory from "./FieldFactory";

/**
 * These are the complex objects that need to create new childs.
 */
export class ObjectField {
  childDef = [];
  /**
   * How many times we've added an element chunk
   */
  elementAddCount = 0;
  /**
   * How many elements we're adding each time
   */
  elementChunkCount = 0;

  constructor(props) {
    this.props = props;
    this.addElements = this.addElements.bind(this);
    this.deleteElements = this.deleteElements.bind(this);
    if (Array.isArray(this.props.fieldDef.fields)) {
      // let's copy the reference to keep a clear reference in memory.
      this.childDef = JSON.parse(JSON.stringify(this.props.fieldDef.fields));
      this.elementChunkCount = this.props.fieldDef.fields.length;
    }
  }

  getJsx() {
    return (
      <ActionGroup
        fieldid={this.props.ids.fieldGroupId}
        key={this.props.ids.fieldGroupKey}
      >
        <Button
          variant="secondary"
          id={this.props.ids.fieldId}
          key={this.props.ids.fieldKey}
          fieldnumber={this.props.fieldNumber}
          onClick={this.addElements}
        >
          Add new {this.props.fieldDef.label}
        </Button>
        <Button
          variant="secondary"
          id={this.props.ids.fieldId + 1}
          key={this.props.ids.fieldKey + 1}
          fieldnumber={this.props.fieldNumber}
          onClick={this.deleteElements}
          disabled={this.elementAddCount == 0}
        >
          Delete last {this.props.fieldDef.label}
        </Button>
      </ActionGroup>
    );
  }

  addElements() {
    var children = [];
    if (Array.isArray(this.childDef) && this.childDef.length > 0) {
      children.push(
        ...FieldFactory.newInstances(
          JSON.parse(JSON.stringify(this.childDef)),
          this.props.jsonSchema,
          this.props.pageNumber,
          this.props.page
        )
      );
    }
    //debugger;
    this.props.page.addElements(
      1 + this.elementAddCount * this.elementChunkCount,
      children,
      this.props.ids.fieldGroupId,
      this.elementAddCount,
      this.props.fieldDef.jsonPath
    );
    this.elementAddCount++;
  }

  deleteElements() {
    if (this.elementAddCount > 0) {
      this.props.page.deleteElements(
        1 + this.elementChunkCount * (this.elementAddCount - 1),
        this.elementChunkCount,
        this.props.ids.fieldGroupId
      );
      this.elementAddCount--;
    }
  }
}
