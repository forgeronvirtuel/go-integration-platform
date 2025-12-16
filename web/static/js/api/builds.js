// API calls for Builds

const BuildsAPI = {
  /**
   * Get all builds
   * @returns {Promise<Object>} Response with builds array
   */
  async getAll() {
    console.log("[BuildsAPI] Fetching all builds");
    const response = await fetch(`${mainAPIURL}/builds`);
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
    console.log("[BuildsAPI] Fetching build by ID:", id);
    const response = await fetch(`${mainAPIURL}/builds/${id}`);
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
    console.log("[BuildsAPI] Fetching builds for project ID:", projectId);
    const response = await fetch(`${mainAPIURL}/builds/project/${projectId}`);
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
    console.log("[BuildsAPI] Creating build with payload:", payload);
    const response = await fetch(`${mainAPIURL}/builds/`, {
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
    return `${mainAPIURL}/builds/${id}/download`;
  },
};
