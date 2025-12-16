// API calls for Runners

const RunnersAPI = {
  /**
   * Get all runners
   * @param {string} [status] - Optional status filter (ONLINE, OFFLINE, DRAINING)
   * @returns {Promise<Object>} Response with runners array
   */
  async getAll(status = null) {
    const url = status
      ? `${mainAPIURL}/runners?status=${status}`
      : `${mainAPIURL}/runners`;

    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  },

  /**
   * Get a runner by ID
   * @param {number} id - Runner ID
   * @returns {Promise<Object>} Runner object
   */
  async getById(id) {
    const response = await fetch(`${mainAPIURL}/runners/${id}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  },

  /**
   * Update runner status
   * @param {number} id - Runner ID
   * @param {string} status - New status (ONLINE, OFFLINE, DRAINING)
   * @returns {Promise<Object>} Updated runner
   */
  async updateStatus(id, status) {
    const response = await fetch(`${mainAPIURL}/runners/${id}/status`, {
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
   * Update runner labels
   * @param {number} id - Runner ID
   * @param {Object} labels - Labels object
   * @returns {Promise<Object>} Updated runner
   */
  async updateLabels(id, labels) {
    const response = await fetch(`${mainAPIURL}/runners/${id}/labels`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ labels }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
  },

  /**
   * Delete a runner
   * @param {number} id - Runner ID
   * @returns {Promise<void>}
   */
  async delete(id) {
    const response = await fetch(`${mainAPIURL}/runners/${id}`, {
      method: "DELETE",
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },
};
