import React, { Component } from "react";
import ElementFactory from "../element-component/ElementFactory";
import { Form, Title } from "@patternfly/react-core";

/**
 * The Page component to handle each element individually.
 */
export default class Page extends Component {
  /**
   * Default constructor for the PageComponent.
   *
   * @param {*} props { pageDef, jsonSchema, pageNumber }
   */
  constructor(props) {
    super(props);
    this.state = {
      elements: []
    };
    this.formId = "form-page-" + this.props.pageNumber;
  }

  loadPageChildren() {
    var elements = ElementFactory.newInstances(
      this.props.pageDef.fields,
      this.props.pageDef.buttons,
      this.props.jsonSchema,
      this.props.pageNumber,
      this
    );

    this.setState({
      elements: elements
    });
  }

  /**
   * Adds a new element to the specific position at the Page and re-render the DOM.
   * @param {int} startIndex
   * @param {Element} element
   */
  addElements(startIndex, newElements, objectkey) {
    this.state.elements.forEach((element, i) => {
      if (element.props != undefined && element.props.ids != undefined) {
        if (element.props.ids.fieldGroupId === objectkey) {
          startIndex = startIndex + i;
        }
      }
    });

    if (Array.isArray(newElements)) {
      var elements = this.state.elements;
      newElements.forEach((element, count) => {
        // update the json state
        this.props.pageDef.fields.splice(
          startIndex + count - 1,
          0,
          JSON.parse(JSON.stringify(element.props.fieldDef))
        );
        // add the elements dynamically
        elements.splice(startIndex + count, 0, element);
      });

      this.setState({ elements: elements });
    } else {
      throw new Error(
        "When adding new elements to the page, please use an Array. Got: ",
        newElements
      );
    }
  }

  /**
   * Removes elements from the Page.
   *
   * @param {int} startIndex
   * @param {int} elementCount
   */
  deleteElements(startIndex, elementCount, objectkey) {
    this.state.elements.forEach((element, i) => {
      // console.log(element.page.props.key);
      if (element.props != undefined && element.props.ids != undefined) {
        if (element.props.ids.fieldGroupId === objectkey) {
          startIndex = startIndex + i;
        }
      }
    });
    var elements = this.state.elements;
    this.props.pageDef.fields.splice(startIndex - 1, elementCount);
    elements.splice(startIndex, elementCount);
    this.setState({ elements: elements });
  }

  componentDidMount() {
    this.loadPageChildren();
  }

  render() {
    return (
      <Form id={"form-page-" + this.props.pageNumber}>
        <Title headingLevel="h1" size="3xl">
          {this.props.pageDef.label}
        </Title>
        {this.state.elements.map(element => {
          return element.getJsx();
        })}
      </Form>
    );
  }
}
