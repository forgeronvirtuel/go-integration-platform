// API calls for Builds

const BuildsAPI = {
  /**
   * Get all builds
   * @returns {Promise<Object>} Response with builds array
   */
  async getAll() {
    const response = await fetch("/v1/api/builds");
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  },

  /**
   * Get a build by ID
   * @param {number} id - Build ID
   * @returns {Promise<Object>} Build object
   */
  async getById(id) {
    const response = await fetch(`/v1/api/builds/${id}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  },

  /**
   * Get builds by project ID
   * @param {number} projectId - Project ID
   * @returns {Promise<Array>} Array of builds
   */
  async getByProjectId(projectId) {
    const response = await fetch(`/v1/api/builds/project/${projectId}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  },

  /**
   * Create a new build
   * @param {Object} payload - Build data
   * @param {number} payload.project_id - Project ID
   * @param {string} payload.branch - Branch name
   * @returns {Promise<Object>} Created build
   */
  async create(payload) {
    const response = await fetch("/v1/api/builds/", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to create build");
    }

    return await response.json();
  },

  /**
   * Download a build binary
   * @param {number} id - Build ID
   * @returns {string} Download URL
   */
  getDownloadUrl(id) {
    return `/v1/api/builds/${id}/download`;
  },
};
