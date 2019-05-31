import React from "react";
import { DropdownField } from "./DropdownField";
import { TextAreaField } from "./TextAreaField";
import { RadioButtonField } from "./RadioButtonField";
import { EmailField } from "./EmailField";
import { UrlField } from "./UrlField";
import { PasswordField } from "./PasswordField";
import { CheckboxField } from "./CheckboxField";
import { SectionField } from "./SectionField";
import { DefaultTextField } from "./DefaultTextField";
import { SectionRadioField } from "./SectionRadioField";
import { ObjectField } from "./ObjectField";
import { FieldUtils } from "./FieldUtils";
import { FieldGroupField } from "./FieldGroupField";
import { IntegerField } from "./IntegerField";

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
  sectionRadio: "section_radio",
  fieldGroup: "fieldGroup",
  integer: "integer"
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
    parentid,
    grandParentId
  ) {
    let fieldReference;
    let elementJsx;
    let props = {
      page: page,
      fieldDef: fieldDef,
      fieldNumber: fieldNumber,
      pageNumber: pageNumber,
      jsonSchema: jsonSchema,
      ids: FieldUtils.generateIds(
        pageNumber,
        fieldNumber,
        fieldDef.label,
        parentid,
        grandParentId
      ),
      parentid: parentid,
      grandParentId: grandParentId
    };

    switch (fieldDef.type) {
      case FIELD_TYPE.dropdown:
        elementJsx = (
          <DropdownField
            props={props}
            key={fieldNumber}
            fieldNumber={fieldNumber}
            fieldDef={fieldDef}
            pageNumber={pageNumber}
            jsonSchema={jsonSchema}
            ids={props.ids}
            parentid={parentid}
          />
        );
        break;
      case FIELD_TYPE.textArea:
        elementJsx = (
          <TextAreaField
            props={props}
            key={fieldNumber}
            fieldNumber={fieldNumber}
            fieldDef={fieldDef}
            pageNumber={pageNumber}
            jsonSchema={jsonSchema}
            ids={props.ids}
            parentid={parentid}
          />
        );
        break;
      case FIELD_TYPE.radioButton:
        fieldReference = new RadioButtonField(props);
        elementJsx = fieldReference.getJsx();
        break;
      case FIELD_TYPE.email:
        elementJsx = (
          <EmailField
            props={props}
            key={fieldNumber}
            fieldNumber={fieldNumber}
            fieldDef={fieldDef}
            pageNumber={pageNumber}
            jsonSchema={jsonSchema}
            ids={props.ids}
            parentid={parentid}
          />
        );
        break;
      case FIELD_TYPE.url:
        elementJsx = (
          <UrlField
            props={props}
            key={fieldNumber}
            fieldNumber={fieldNumber}
            fieldDef={fieldDef}
            pageNumber={pageNumber}
            jsonSchema={jsonSchema}
            ids={props.ids}
            parentid={parentid}
          />
        );
        break;
      case FIELD_TYPE.password:
        elementJsx = (
          <PasswordField
            props={props}
            key={fieldNumber}
            fieldNumber={fieldNumber}
            fieldDef={fieldDef}
            pageNumber={pageNumber}
            jsonSchema={jsonSchema}
            ids={props.ids}
            parentid={parentid}
          />
        );
        break;
      case FIELD_TYPE.checkbox:
        elementJsx = (
          <CheckboxField
            props={props}
            key={fieldNumber}
            fieldNumber={fieldNumber}
            fieldDef={fieldDef}
            pageNumber={pageNumber}
            jsonSchema={jsonSchema}
            ids={props.ids}
            parentid={parentid}
          />
        );
        break;
      case FIELD_TYPE.section:
        fieldReference = new SectionField(props);
        elementJsx = fieldReference.getJsx();
        break;
      case FIELD_TYPE.sectionRadio:
        fieldReference = new SectionRadioField(props);
        elementJsx = fieldReference.getJsx();
        break;
      case FIELD_TYPE.object:
        if (props.parentid === undefined) {
          props.parentid = -1;
        }
        fieldReference = new ObjectField(props);
        elementJsx = fieldReference.getJsx();
        break;
      case FIELD_TYPE.fieldGroup:
        fieldReference = new FieldGroupField(props);
        elementJsx = fieldReference.getJsx();
        break;
      case FIELD_TYPE.integer:
        elementJsx = (
          <IntegerField
            props={props}
            key={fieldNumber}
            fieldNumber={fieldNumber}
            fieldDef={fieldDef}
            pageNumber={pageNumber}
            jsonSchema={jsonSchema}
            ids={props.ids}
            parentid={parentid}
          />
        );
        break;
      default:
        elementJsx = (
          <DefaultTextField
            props={props}
            key={fieldNumber}
            fieldNumber={fieldNumber}
            fieldDef={fieldDef}
            pageNumber={pageNumber}
            jsonSchema={jsonSchema}
            ids={props.ids}
            parentid={parentid}
          />
        );
        break;
    }

    return elementJsx;
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
