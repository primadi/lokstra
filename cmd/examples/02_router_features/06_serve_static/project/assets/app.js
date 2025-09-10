// Project override app.js
console.log("Project override JavaScript loaded - HIGHEST PRIORITY");

// This file will be served first if it exists in project/assets
function projectOverrideInit() {
  console.log("Project override initialization complete");

  // Override framework behavior
  document.body.style.backgroundColor = "#e8f4fd";
}
