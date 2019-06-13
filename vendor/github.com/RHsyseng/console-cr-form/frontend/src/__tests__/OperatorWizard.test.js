import React from "react";
import ReactDOM from "react-dom";
import OperatorWizard from "../component/operator-wizard/OperatorWizard";
import fs from "fs";

describe("Operator Wizard", () => {
  afterEach(() => {
    fetch.resetMocks();
  });

  it("renders without crashing", () => {
    fetch.mockResponseOnce(readFile("full-form.json"));
    fetch.mockResponseOnce(readFile("full-schema.json"));
    fetch.mockResponseOnce(
      JSON.stringify({ apiVersion: "testApi", kind: "testKind" })
    );
    const div = document.createElement("div");
    ReactDOM.render(<OperatorWizard />, div);
    expect(fetch.mock.calls.length).toEqual(3);
    expect(fetch.mock.calls[0][0]).toEqual("/api/spec");
    expect(fetch.mock.calls[1][0]).toEqual("/api/form");
    expect(fetch.mock.calls[2][0]).toEqual("/api/schema");
  });
});

function readFile(fileName) {
  return fs.readFileSync("../test/examples/" + fileName);
}
