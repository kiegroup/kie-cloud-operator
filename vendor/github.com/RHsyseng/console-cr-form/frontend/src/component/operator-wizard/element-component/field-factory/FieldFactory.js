import { DropdownField } from "./DropdownField";
import { TextAreaField } from "./TextAreaField";
import { RadioButtonField } from "./RadioButtonField";
import { EmailField } from "./EmailField";
import { UrlField } from "./UrlField";
import { PasswordField } from "./PasswordField";
import { CheckboxField } from "./CheckboxField";
import { SeparateDivField } from "./SeparateDivField";
import { SectionField } from "./SectionField";
import { DefaultTextField } from "./DefaultTextField";
import { SectionRadioField } from "./SectionRadioField";
import { ObjectField } from "./ObjectField";
import { FieldUtils } from "./FieldUtils";

export const FIELD_TYPE = {
  dropdown: "dropDown",
  textArea: "textArea",
  radioButton: "radioButton",
  object: "object",
  email: "email",
  url: "url",
  password: "password",
  checkbox: "checkbox",
  seperateObjDiv: "seperateObjDiv",
  section: "section",
  text: "text",
  sectionRadio: "section_radio"
};

export default class FieldFactory {
  /**
   * Creates a single instance of a field
   */
  static newInstance(
    fieldDef,
    fieldNumber,
    pageNumber,
    jsonSchema,
    page,
    parentid
  ) {
    var fieldReference;
    var props = {
      page: page,
      fieldDef: fieldDef,
      fieldNumber: fieldNumber,
      pageNumber: pageNumber,
      jsonSchema: jsonSchema,
      ids: FieldUtils.generateIds(
        pageNumber,
        fieldNumber,
        fieldDef.label,
        parentid
      ),
      parentid: parentid
    };
    //TODO: rethink when we have the time
    switch (fieldDef.type) {
      case FIELD_TYPE.dropdown:
        fieldReference = new DropdownField(props);
        break;
      case FIELD_TYPE.textArea:
        fieldReference = new TextAreaField(props);
        break;
      case FIELD_TYPE.radioButton:
        fieldReference = new RadioButtonField(props);
        break;
      case FIELD_TYPE.email:
        fieldReference = new EmailField(props);
        break;
      case FIELD_TYPE.url:
        fieldReference = new UrlField(props);
        break;
      case FIELD_TYPE.password:
        fieldReference = new PasswordField(props);
        break;
      case FIELD_TYPE.checkbox:
        fieldReference = new CheckboxField(props);
        break;
      case FIELD_TYPE.seperateObjDiv:
        fieldReference = new SeparateDivField(props);
        break;
      case FIELD_TYPE.section:
        fieldReference = new SectionField(props);
        break;
      case FIELD_TYPE.sectionRadio:
        fieldReference = new SectionRadioField(props);
        break;
      case FIELD_TYPE.object:
        if (props.parentid === undefined) {
          props.parentid = -1;
        }

        fieldReference = new ObjectField(props);
        break;
      default:
        fieldReference = new DefaultTextField(props);
    }

    return fieldReference;
  }

  /**
   * Creates all instances based on a field array
   */
  static newInstances(fieldDefs, jsonSchema, pageNumber, page, parentid) {
    var children = [];
    if (fieldDefs !== undefined && fieldDefs !== null && fieldDefs !== "") {
      fieldDefs.forEach((field, fieldNumber) => {
        var fieldGenerator = FieldFactory.newInstance(
          field,
          fieldNumber,
          pageNumber,
          jsonSchema,
          page,
          parentid
        );
        if (fieldGenerator !== null) {
          children.push(fieldGenerator);
        }
      });
    }
    return children;
  }

  /**
   * Same as newInstances, but return an array of instances in the form of JSX.
   */
  static newInstancesAsJsx(fieldDefs, jsonSchema, pageNumber, page, parentId) {
    var children = FieldFactory.newInstances(
      fieldDefs,
      jsonSchema,
      pageNumber,
      page,
      parentId
    );
    var childrenJsx = [];
    children.forEach(child => {
      childrenJsx.push(child.getJsx());
    });
    return childrenJsx;
  }
}
