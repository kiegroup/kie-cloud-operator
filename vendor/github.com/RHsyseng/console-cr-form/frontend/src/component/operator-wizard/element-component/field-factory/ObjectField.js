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
   * (Const) How many elements we're adding each time
   */
  elementChunkCount = 0;
  /**
   * Min fixed elements added to the component
   */
  minElements = [];
  /**
   * Max chunk of elements that could be added to this object.
   */
  maxElementsSize = 0;
  /**
   * Min chunk of elements that could be added to this object.
   */
  minElementsSize = 0;
  constructor(props) {
    this.props = props;
    this.addElements = this.addElements.bind(this);
    this.deleteElements = this.deleteElements.bind(this);
    if (Array.isArray(this.props.fieldDef.fields)) {
      this.minElementsSize =
        this.props.fieldDef.min === undefined
          ? 0
          : parseInt(this.props.fieldDef.min);
      this.maxElementsSize =
        this.props.fieldDef.max === undefined
          ? 0
          : parseInt(this.props.fieldDef.max);
      // let's copy the reference to keep a clear reference in memory.
      this.childDef = JSON.parse(JSON.stringify(this.props.fieldDef.fields));
      this.elementChunkCount = this.props.fieldDef.fields.length;
      this.elementAddCount =
        this.props.fieldDef.elementAddCount === undefined
          ? 0
          : parseInt(this.props.fieldDef.elementAddCount);
      this.minElements = this.addMinElements();
    }
  }

  getJsx() {
    return (
      <div key={"div-" + this.props.ids.fieldGroupKey}>
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
        {this.minElements}
      </div>
    );
  }

  createChildrenChunk() {
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
    return children;
  }

  addMinElements() {
    var elements = [];
    for (let index = 0; index < this.minElementsSize; index++) {
      var children = this.createChildrenChunk();
      children.forEach(child => {
        elements.push(child.getJsx());
      });
    }
    return elements;
  }

  addElements() {
    if (this.maxElementsSize > this.elementAddCount + this.minElementsSize) {
      var children = this.createChildrenChunk();
      this.props.page.addElements(
        1 + this.elementAddCount * this.elementChunkCount,
        children,
        this.props.ids.fieldGroupId
      );
      this.elementAddCount++;
      this.setElementAddCountState(this.elementAddCount);
    }
  }

  deleteElements() {
    if (this.elementAddCount > 0) {
      this.props.page.deleteElements(
        1 + this.elementChunkCount * (this.elementAddCount - 1),
        this.elementChunkCount,
        this.props.ids.fieldGroupId
      );
      this.elementAddCount--;
      this.setElementAddCountState(this.elementAddCount);
    }
  }

  /**
   * Preserve the element add count so we can restore its state between the wizard navigation
   */
  setElementAddCountState(count) {
    this.props.fieldDef.elementAddCount = count;
  }
}
