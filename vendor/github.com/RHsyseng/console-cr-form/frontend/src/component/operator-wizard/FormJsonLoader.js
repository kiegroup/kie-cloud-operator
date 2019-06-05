import { BACKEND_URL } from "../common/GuiConstants";

export const loadJsonForm = fetch(BACKEND_URL + "/form", {
  headers: {
    "Content-Type": "application/json"
  },
  credentials: "same-origin"
})
  .then(res => res.json())
  .catch(error => {
    console.error("Unable to load JSON Form: ", error);
  });

export const loadJsonSchema = fetch(BACKEND_URL + "/schema", {
  headers: {
    "Content-Type": "application/json"
  },
  credentials: "same-origin"
})
  .then(res => res.json())
  .catch(error => {
    console.error("Unable to load JSON Schema: ", error);
  });

export const loadJsonSpec = () =>
  fetch(BACKEND_URL + "/spec", {
    headers: {
      "Content-Type": "application/json"
    },
    credentials: "same-origin"
  })
    .then(res => res.json())
    .catch(error => {
      console.error("Unable to load JSON Schema: ", error);
    });
