// API calls for Deployments

const DeploymentsAPI = {
  /**
   * Get all deployments
   * @returns {Promise<Object>} Response with deployments array
   */
  async getAll() {
    const response = await fetch("/v1/deployments");
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  },

  /**
   * Get a deployment by ID
   * @param {number} id - Deployment ID
   * @returns {Promise<Object>} Deployment object
   */
  async getById(id) {
    const response = await fetch(`/v1/deployments/${id}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  },

  /**
   * Create a new deployment
   * @param {Object} payload - Deployment data
   * @param {number} payload.build_id - Build ID
   * @param {number} [payload.runner_id] - Optional runner ID
   * @returns {Promise<Object>} Created deployment
   */
  async create(payload) {
    const response = await fetch("/v1/deployments", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to create deployment");
    }

    return await response.json();
  },

  /**
   * Update deployment status
   * @param {number} id - Deployment ID
   * @param {string} status - New status (pending, deploying, deployed, failed)
   * @returns {Promise<Object>} Updated deployment
   */
  async updateStatus(id, status) {
    const response = await fetch(`/v1/deployments/${id}/status`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ status }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
  },

  /**
   * Update deployment runner
   * @param {number} id - Deployment ID
   * @param {number|null} runnerId - Runner ID or null to unassign
   * @returns {Promise<Object>} Updated deployment
   */
  async updateRunner(id, runnerId) {
    const response = await fetch(`/v1/deployments/${id}/runner`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ runner_id: runnerId }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to update runner");
    }

    return await response.json();
  },

  /**
   * Execute a deployment
   * @param {number} id - Deployment ID
   * @returns {Promise<Object>} Execution result
   */
  async execute(id) {
    const response = await fetch(`/v1/deployments/${id}/execute`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to execute deployment");
    }

    return await response.json();
  },

  /**
   * Delete a deployment
   * @param {number} id - Deployment ID
   * @returns {Promise<void>}
   */
  async delete(id) {
    const response = await fetch(`/v1/deployments/${id}`, {
      method: "DELETE",
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },
};
