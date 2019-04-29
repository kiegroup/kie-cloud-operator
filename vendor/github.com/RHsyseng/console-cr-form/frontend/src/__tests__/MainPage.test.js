import React from "react";
import ReactDOM from "react-dom";
import MainPage from "../component/MainPage";

describe("Main Page", () => {
  it("renders without crashing", () => {
    const div = document.createElement("div");
    ReactDOM.render(<MainPage />, div);
  });
});
