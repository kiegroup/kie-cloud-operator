import { ButtonGroup } from "./button-element/ButtonGroup";
import FieldFactory from "./field-factory/FieldFactory";

export default class ElementFactory {
  /**
   * Creates an array of element instances to be populated on each step
   * @param {*} fieldDefs the field defined by the JSON loaded into mem
   * @param {*} buttonDefs the button defined by the JSON loaded into mem
   * @param {*} jsonSchema the JSON Schema loaded into mem TODO: try to remove this since only the dropDown element requires it
   * @param {int} pageNumber the Page/Step number
   * @param {Page} page the page reference
   */
  static newInstances(fieldDefs, buttonDefs, jsonSchema, pageNumber, page) {
    const children = [];
    children.push(new ButtonGroup(buttonDefs, pageNumber, page));
    children.push(
      ...FieldFactory.newInstances(fieldDefs, jsonSchema, pageNumber, page)
    );

    return children;
  }
}
