import { USE_MOCK_DATA } from "../common/GuiConstants";

//TODO: dynamically load
import { MockupData_JSON, MockupData_JSON_SCHEMA } from "../common/MockupData";

export class FormJsonLoader {
  passedInJsonForm = {};
  jsonForm = {};
  jsonSchema = {};

  constructor({ elementIdJson, elementIdJsonSchema }) {
    this.elementIdJson = elementIdJson;
    this.elementIdJsonSchema = elementIdJsonSchema;
    //TODO: turn it into promises
    this.loadJson(USE_MOCK_DATA);
  }

  loadJson(useMockData) {
    if (useMockData) {
      this.passedInJsonForm = MockupData_JSON;
      this.jsonSchema = MockupData_JSON_SCHEMA;
      console.log("Loaded mock data into memmory");
    } else {
      //TODO: refactor to not access DOM directly
      this.passedInJsonForm = document.getElementById(
        this.elementIdJson
      ).innerHTML;
      this.jsonSchema = document.getElementById(
        this.elementIdJsonSchema
      ).innerHTML;
    }

    this.jsonForm = JSON.parse(JSON.stringify(this.passedInJsonForm));
  }
}
