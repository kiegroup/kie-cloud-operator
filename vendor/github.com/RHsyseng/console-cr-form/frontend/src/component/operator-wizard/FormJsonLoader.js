import { BACKEND_URL } from "../common/GuiConstants";

class FormJsonLoader {
  loadJsonForm() {
    return fetch(BACKEND_URL + "/form", {
      headers: {
        "Content-Type": "application/json"
      },
      credentials: "same-origin"
    })
      .then(res => res.json())
      .catch(error => {
        console.error("Unable to load JSON Form: ", error);
      });
  }

  loadJsonSchema() {
    return fetch(BACKEND_URL + "/schema", {
      headers: {
        "Content-Type": "application/json"
      },
      credentials: "same-origin"
    })
      .then(res => res.json())
      .catch(error => {
        console.error("Unable to load JSON Schema: ", error);
      });
  }

  loadJsonSpec() {
    return fetch(BACKEND_URL + "/spec", {
      headers: {
        "Content-Type": "application/json"
      },
      credentials: "same-origin"
    })
      .then(res => res.json())
      .catch(error => {
        console.error("Unable to load JSON Schema: ", error);
      });
  }
}

export default new FormJsonLoader();
