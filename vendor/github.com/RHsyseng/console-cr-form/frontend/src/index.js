import React from "react";
import ReactDOM from "react-dom";
import "@patternfly/react-core/dist/styles/base.css";
import MainPage from "./component/Main";

document.addEventListener("DOMContentLoaded", function() {
  ReactDOM.render(
    React.createElement(MainPage),
    document.getElementById("mount")
  );
});
