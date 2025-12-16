// API calls for Projects

const ProjectsAPI = {
  /**
   * Get all projects
   * @returns {Promise<Object>} Response with projects array and count
   */
  async getAll() {
    const response = await fetch(`${mainAPIURL}/projects`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  },

  /**
   * Get a project by ID
   * @param {number} id - Project ID
   * @returns {Promise<Object>} Project object
   */
  async getById(id) {
    const response = await fetch(`${mainAPIURL}/projects/${id}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  },

  /**
   * Create a new project
   * @param {Object} payload - Project data
   * @param {string} payload.name - Project name
   * @param {string} payload.repo_url - Repository URL
   * @param {string} payload.branch - Branch name
   * @param {string} [payload.subdir] - Optional subdirectory
   * @returns {Promise<Object>} Created project
   */
  async create(payload) {
    const response = await fetch(`${mainAPIURL}/projects`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to create project");
    }

    return await response.json();
  },

  /**
   * Update a project
   * @param {number} id - Project ID
   * @param {Object} payload - Project data to update
   * @returns {Promise<Object>} Updated project
   */
  async update(id, payload) {
    const response = await fetch(`${mainAPIURL}/projects/${id}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to update project");
    }

    return await response.json();
  },

  /**
   * Delete a project
   * @param {number} id - Project ID
   * @returns {Promise<void>}
   */
  async delete(id) {
    const response = await fetch(`${mainAPIURL}/projects/${id}`, {
      method: "DELETE",
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },
};
