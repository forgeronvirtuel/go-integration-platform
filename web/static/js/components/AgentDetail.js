// Composant AgentDetail
const { useState, useEffect } = React;

function AgentDetail({ agent, onMessage, onBack }) {
  const [agentData, setAgentData] = useState(agent);
  const [loading, setLoading] = useState(false);
  const [editingLabels, setEditingLabels] = useState(false);
  const [newLabels, setNewLabels] = useState({});

  console.log("🔍 [AgentDetail] Component mounted with agent:", agent);

  const loadAgentDetails = async () => {
    try {
      console.log("🔍 [AgentDetail] Loading agent details for ID:", agent.id);
      setLoading(true);

      const response = await fetch(`/v1/api/agents/${agent.id}`);

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      console.log("🔍 [AgentDetail] Agent details loaded:", data);

      setAgentData(data);
      setLoading(false);
    } catch (error) {
      console.error("🔍 [AgentDetail] Error loading agent details:", error);
      onMessage(`Erreur lors du chargement des détails: ${error.message}`);
      setLoading(false);
    }
  };

  const updateStatus = async (newStatus) => {
    try {
      console.log("🔍 [AgentDetail] Updating status to:", newStatus);

      const response = await fetch(`/v1/api/agents/${agent.id}/status`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ status: newStatus }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      console.log("🔍 [AgentDetail] Status updated:", data);

      setAgentData(data);
      onMessage(`Statut mis à jour: ${newStatus}`);
    } catch (error) {
      console.error("🔍 [AgentDetail] Error updating status:", error);
      onMessage(`Erreur lors de la mise à jour du statut: ${error.message}`);
    }
  };

  const updateLabels = async () => {
    try {
      console.log("🔍 [AgentDetail] Updating labels:", newLabels);

      const response = await fetch(`/v1/api/agents/${agent.id}/labels`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ labels: newLabels }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      console.log("🔍 [AgentDetail] Labels updated:", data);

      setAgentData(data);
      setEditingLabels(false);
      onMessage("Labels mis à jour avec succès");
    } catch (error) {
      console.error("🔍 [AgentDetail] Error updating labels:", error);
      onMessage(`Erreur lors de la mise à jour des labels: ${error.message}`);
    }
  };

  const deleteAgent = async () => {
    if (
      !confirm(
        `Êtes-vous sûr de vouloir supprimer l'agent "${agentData.name}" ?`
      )
    ) {
      return;
    }

    try {
      console.log("🔍 [AgentDetail] Deleting agent:", agent.id);

      const response = await fetch(`/v1/api/agents/${agent.id}`, {
        method: "DELETE",
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      console.log("🔍 [AgentDetail] Agent deleted successfully");
      onMessage("Agent supprimé avec succès");
      onBack();
    } catch (error) {
      console.error("🔍 [AgentDetail] Error deleting agent:", error);
      onMessage(`Erreur lors de la suppression: ${error.message}`);
    }
  };

  useEffect(() => {
    loadAgentDetails();
    // Rafraîchir toutes les 5 secondes
    const interval = setInterval(loadAgentDetails, 5000);
    return () => clearInterval(interval);
  }, [agent.id]);

  useEffect(() => {
    setNewLabels(agentData.labels || {});
  }, [agentData.labels]);

  const getStatusBadge = (status) => {
    const statusColors = {
      ONLINE: "bg-green-100 text-green-800 border-green-300",
      OFFLINE: "bg-gray-100 text-gray-800 border-gray-300",
      DRAINING: "bg-yellow-100 text-yellow-800 border-yellow-300",
    };

    return (
      <span
        className={`px-4 py-2 rounded-full text-lg font-semibold border ${
          statusColors[status] || "bg-gray-100 text-gray-800"
        }`}
      >
        {status}
      </span>
    );
  };

  const formatDate = (dateString) => {
    if (!dateString) return "Jamais";
    const date = new Date(dateString);
    return new Intl.DateTimeFormat("fr-FR", {
      dateStyle: "full",
      timeStyle: "long",
    }).format(date);
  };

  const getTimeSince = (dateString) => {
    if (!dateString) return "Jamais";
    const now = new Date();
    const date = new Date(dateString);
    const seconds = Math.floor((now - date) / 1000);

    if (seconds < 60)
      return `il y a ${seconds} seconde${seconds !== 1 ? "s" : ""}`;
    if (seconds < 3600)
      return `il y a ${Math.floor(seconds / 60)} minute${
        Math.floor(seconds / 60) !== 1 ? "s" : ""
      }`;
    if (seconds < 86400)
      return `il y a ${Math.floor(seconds / 3600)} heure${
        Math.floor(seconds / 3600) !== 1 ? "s" : ""
      }`;
    return `il y a ${Math.floor(seconds / 86400)} jour${
      Math.floor(seconds / 86400) !== 1 ? "s" : ""
    }`;
  };

  if (loading && !agentData.name) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center space-x-4">
        <button
          onClick={onBack}
          className="text-blue-600 hover:text-blue-800 font-medium"
        >
          ← Retour à la liste
        </button>
      </div>

      <div className="bg-white border border-gray-200 rounded-lg p-8">
        <div className="flex justify-between items-start mb-6">
          <div>
            <h2 className="text-3xl font-bold text-gray-800 mb-2">
              🤖 {agentData.name}
            </h2>
            <p className="text-gray-600">ID: #{agentData.id}</p>
          </div>
          {getStatusBadge(agentData.status)}
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          <div className="space-y-4">
            <div className="bg-gray-50 p-4 rounded-lg">
              <h3 className="text-sm font-semibold text-gray-700 mb-2">
                📅 Date de création
              </h3>
              <p className="text-gray-800">
                {formatDate(agentData.created_at)}
              </p>
            </div>

            <div className="bg-gray-50 p-4 rounded-lg">
              <h3 className="text-sm font-semibold text-gray-700 mb-2">
                💓 Dernier heartbeat
              </h3>
              <p className="text-gray-800">
                {agentData.last_seen_at ? (
                  <>
                    <span className="text-blue-600 font-medium">
                      {getTimeSince(agentData.last_seen_at)}
                    </span>
                    <br />
                    <span className="text-sm text-gray-600">
                      {formatDate(agentData.last_seen_at)}
                    </span>
                  </>
                ) : (
                  <span className="text-gray-500">Aucun heartbeat reçu</span>
                )}
              </p>
            </div>
          </div>

          <div className="space-y-4">
            <div className="bg-gray-50 p-4 rounded-lg">
              <h3 className="text-sm font-semibold text-gray-700 mb-3">
                ⚙️ Actions
              </h3>
              <div className="space-y-2">
                {agentData.status !== "ONLINE" && (
                  <button
                    onClick={() => updateStatus("ONLINE")}
                    className="w-full bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors"
                  >
                    ✅ Mettre en ligne
                  </button>
                )}
                {agentData.status !== "OFFLINE" && (
                  <button
                    onClick={() => updateStatus("OFFLINE")}
                    className="w-full bg-gray-600 text-white px-4 py-2 rounded-lg hover:bg-gray-700 transition-colors"
                  >
                    🚫 Mettre hors ligne
                  </button>
                )}
                {agentData.status !== "DRAINING" && (
                  <button
                    onClick={() => updateStatus("DRAINING")}
                    className="w-full bg-yellow-600 text-white px-4 py-2 rounded-lg hover:bg-yellow-700 transition-colors"
                  >
                    ⏸️ Mode DRAINING
                  </button>
                )}
                <button
                  onClick={deleteAgent}
                  className="w-full bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors"
                >
                  🗑️ Supprimer l'agent
                </button>
              </div>
            </div>
          </div>
        </div>

        <div className="border-t pt-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-xl font-semibold text-gray-800">🏷️ Labels</h3>
            <button
              onClick={() => setEditingLabels(!editingLabels)}
              className="text-blue-600 hover:text-blue-800 font-medium"
            >
              {editingLabels ? "❌ Annuler" : "✏️ Modifier"}
            </button>
          </div>

          {!editingLabels ? (
            <div>
              {agentData.labels && Object.keys(agentData.labels).length > 0 ? (
                <div className="flex flex-wrap gap-3">
                  {Object.entries(agentData.labels).map(([key, value]) => (
                    <div
                      key={key}
                      className="bg-blue-50 text-blue-700 px-4 py-2 rounded-lg border border-blue-200"
                    >
                      <span className="font-semibold">{key}:</span> {value}
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500">Aucun label défini</p>
              )}
            </div>
          ) : (
            <div className="space-y-4">
              <div className="bg-gray-50 p-4 rounded-lg">
                <p className="text-sm text-gray-600 mb-4">
                  Modifiez les labels ci-dessous. Format: une ligne par label
                  (clé=valeur)
                </p>
                <textarea
                  value={Object.entries(newLabels)
                    .map(([k, v]) => `${k}=${v}`)
                    .join("\n")}
                  onChange={(e) => {
                    const lines = e.target.value.split("\n");
                    const labels = {};
                    lines.forEach((line) => {
                      const [key, ...valueParts] = line.split("=");
                      if (key && valueParts.length > 0) {
                        labels[key.trim()] = valueParts.join("=").trim();
                      }
                    });
                    setNewLabels(labels);
                  }}
                  className="w-full border border-gray-300 rounded-lg p-3 h-32 font-mono text-sm"
                  placeholder="env=production&#10;region=eu-west&#10;capacity=high"
                />
              </div>
              <button
                onClick={updateLabels}
                className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors"
              >
                💾 Sauvegarder les labels
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
