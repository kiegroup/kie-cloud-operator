import React from "react";
import ReactDOM from "react-dom";
import "@patternfly/react-core/dist/styles/base.css";
import Main from "./component/Main";

document.addEventListener("DOMContentLoaded", function() {
  ReactDOM.render(React.createElement(Main), document.getElementById("mount"));
});
