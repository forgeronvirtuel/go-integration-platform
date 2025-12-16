// API Index - Central export for all API modules

/**
 * Main API object that groups all API modules
 */
const API = {
  deployments: DeploymentsAPI,
  builds: BuildsAPI,
  projects: ProjectsAPI,
  runners: RunnersAPI,
};

// Export for use in components
if (typeof module !== "undefined" && module.exports) {
  module.exports = API;
}
