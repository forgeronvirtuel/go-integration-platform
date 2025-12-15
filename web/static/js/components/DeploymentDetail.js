// Composant DeploymentDetail
const { useState, useEffect } = React;

function DeploymentDetail({ deployment, onMessage, onBack }) {
  const [deploymentData, setDeploymentData] = useState(deployment);
  const [build, setBuild] = useState(null);
  const [project, setProject] = useState(null);
  const [agent, setAgent] = useState(null);
  const [availableAgents, setAvailableAgents] = useState([]);
  const [loading, setLoading] = useState(false);
  const [editingAgent, setEditingAgent] = useState(false);
  const [selectedAgentID, setSelectedAgentID] = useState("");

  console.log(
    "🔍 [DeploymentDetail] Component mounted with deployment:",
    deployment
  );

  useEffect(() => {
    loadDeploymentDetails();
    loadAvailableAgents();
    // Rafraîchir toutes les 5 secondes
    const interval = setInterval(loadDeploymentDetails, 5000);
    return () => clearInterval(interval);
  }, [deployment.id]);

  const loadDeploymentDetails = async () => {
    try {
      console.log("🔍 [DeploymentDetail] Loading deployment details...");
      setLoading(true);

      // Charger le déploiement
      const deploymentResponse = await fetch(
        `/v1/deployments/${deployment.id}`
      );
      const deploymentData = await deploymentResponse.json();
      setDeploymentData(deploymentData);

      // Charger le build
      const buildResponse = await fetch(
        `/v1/builds/${deploymentData.build_id}`
      );
      const buildData = await buildResponse.json();
      setBuild(buildData);

      // Charger le projet
      const projectResponse = await fetch(
        `/v1/projects/${buildData.project_id}`
      );
      const projectData = await projectResponse.json();
      setProject(projectData);

      // Charger l'agent si assigné
      if (deploymentData.agent_id) {
        const agentResponse = await fetch(
          `/v1/api/agents/${deploymentData.agent_id}`
        );
        const agentData = await agentResponse.json();
        setAgent(agentData);
      } else {
        setAgent(null);
      }

      setLoading(false);
    } catch (error) {
      console.error("🔍 [DeploymentDetail] Error loading details:", error);
      onMessage(`Erreur lors du chargement: ${error.message}`);
      setLoading(false);
    }
  };

  const loadAvailableAgents = async () => {
    try {
      const response = await fetch("/v1/api/agents?status=ONLINE");
      const data = await response.json();
      setAvailableAgents(data.agents || []);
    } catch (error) {
      console.error("🔍 [DeploymentDetail] Error loading agents:", error);
    }
  };

  const updateStatus = async (newStatus) => {
    try {
      console.log("🔍 [DeploymentDetail] Updating status to:", newStatus);

      const response = await fetch(`/v1/deployments/${deployment.id}/status`, {
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
      setDeploymentData(data);
      onMessage(`Statut mis à jour: ${newStatus}`);
    } catch (error) {
      console.error("🔍 [DeploymentDetail] Error updating status:", error);
      onMessage(`Erreur: ${error.message}`);
    }
  };

  const updateAgent = async () => {
    try {
      console.log("🔍 [DeploymentDetail] Updating agent to:", selectedAgentID);

      const payload = {
        agent_id: selectedAgentID ? parseInt(selectedAgentID) : null,
      };

      const response = await fetch(`/v1/deployments/${deployment.id}/agent`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to update agent");
      }

      const data = await response.json();
      setDeploymentData(data);
      setEditingAgent(false);
      onMessage("Agent mis à jour avec succès");
      loadDeploymentDetails();
    } catch (error) {
      console.error("🔍 [DeploymentDetail] Error updating agent:", error);
      onMessage(`Erreur: ${error.message}`);
    }
  };

  const deleteDeployment = async () => {
    if (
      !confirm(
        `Êtes-vous sûr de vouloir supprimer le déploiement #${deploymentData.id} ?`
      )
    ) {
      return;
    }

    try {
      console.log("🔍 [DeploymentDetail] Deleting deployment:", deployment.id);

      const response = await fetch(`/v1/deployments/${deployment.id}`, {
        method: "DELETE",
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      onMessage("Déploiement supprimé avec succès");
      onBack();
    } catch (error) {
      console.error("🔍 [DeploymentDetail] Error deleting deployment:", error);
      onMessage(`Erreur: ${error.message}`);
    }
  };

  const executeDeployment = async () => {
    try {
      console.log("🔍 [DeploymentDetail] Executing deployment:", deployment.id);

      const response = await fetch(`/v1/deployments/${deployment.id}/execute`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to execute deployment");
      }

      onMessage("Déploiement lancé avec succès");
      loadDeploymentDetails(); // Rafraîchir les données
    } catch (error) {
      console.error("🔍 [DeploymentDetail] Error executing deployment:", error);
      onMessage(`Erreur: ${error.message}`);
    }
  };

  const getStatusBadge = (status) => {
    const statusColors = {
      pending: "bg-yellow-100 text-yellow-800 border-yellow-300",
      deploying: "bg-blue-100 text-blue-800 border-blue-300",
      deployed: "bg-green-100 text-green-800 border-green-300",
      failed: "bg-red-100 text-red-800 border-red-300",
    };

    const statusLabels = {
      pending: "⏳ En attente",
      deploying: "🚀 En cours",
      deployed: "✅ Déployé",
      failed: "❌ Échoué",
    };

    return (
      <span
        className={`px-4 py-2 rounded-full text-lg font-semibold border ${
          statusColors[status] || "bg-gray-100 text-gray-800"
        }`}
      >
        {statusLabels[status] || status}
      </span>
    );
  };

  const formatDate = (dateString) => {
    if (!dateString) return "N/A";
    const date = new Date(dateString);
    return new Intl.DateTimeFormat("fr-FR", {
      dateStyle: "full",
      timeStyle: "long",
    }).format(date);
  };

  if (loading && !deploymentData.id) {
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
              🚀 Déploiement #{deploymentData.id}
            </h2>
            {project && (
              <p className="text-gray-600 text-lg">Projet: {project.name}</p>
            )}
          </div>
          {getStatusBadge(deploymentData.status)}
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          <div className="space-y-4">
            <div className="bg-gray-50 p-4 rounded-lg">
              <h3 className="text-sm font-semibold text-gray-700 mb-2">
                🔨 Build
              </h3>
              {build ? (
                <div className="text-gray-800">
                  <p>
                    <strong>ID:</strong> #{build.id}
                  </p>
                  <p>
                    <strong>Branche:</strong> {build.branch}
                  </p>
                  <p>
                    <strong>Statut:</strong> {build.status}
                  </p>
                </div>
              ) : (
                <p className="text-gray-500">Chargement...</p>
              )}
            </div>

            <div className="bg-gray-50 p-4 rounded-lg">
              <h3 className="text-sm font-semibold text-gray-700 mb-2">
                📅 Dates
              </h3>
              <div className="text-gray-800 space-y-1">
                <p>
                  <strong>Créé:</strong> {formatDate(deploymentData.created_at)}
                </p>
                <p>
                  <strong>Démarré:</strong>{" "}
                  {formatDate(deploymentData.started_at)}
                </p>
                {deploymentData.ended_at && (
                  <p>
                    <strong>Terminé:</strong>{" "}
                    {formatDate(deploymentData.ended_at)}
                  </p>
                )}
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div className="bg-gray-50 p-4 rounded-lg">
              <div className="flex justify-between items-center mb-3">
                <h3 className="text-sm font-semibold text-gray-700">
                  🤖 Agent
                </h3>
                <button
                  onClick={() => {
                    setEditingAgent(!editingAgent);
                    setSelectedAgentID(deploymentData.agent_id || "");
                  }}
                  className="text-blue-600 hover:text-blue-800 text-sm font-medium"
                >
                  {editingAgent ? "❌ Annuler" : "✏️ Modifier"}
                </button>
              </div>

              {!editingAgent ? (
                agent ? (
                  <div className="text-gray-800">
                    <p>
                      <strong>Nom:</strong> {agent.name}
                    </p>
                    <p>
                      <strong>ID:</strong> #{agent.id}
                    </p>
                    <p>
                      <strong>Statut:</strong> {agent.status}
                    </p>
                  </div>
                ) : (
                  <p className="text-gray-500">Aucun agent assigné</p>
                )
              ) : (
                <div className="space-y-3">
                  <select
                    value={selectedAgentID}
                    onChange={(e) => setSelectedAgentID(e.target.value)}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2"
                  >
                    <option value="">-- Aucun agent --</option>
                    {availableAgents.map((a) => (
                      <option key={a.id} value={a.id}>
                        {a.name} (#{a.id})
                      </option>
                    ))}
                  </select>
                  <button
                    onClick={updateAgent}
                    className="w-full bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
                  >
                    💾 Sauvegarder
                  </button>
                </div>
              )}
            </div>

            <div className="bg-gray-50 p-4 rounded-lg">
              <h3 className="text-sm font-semibold text-gray-700 mb-3">
                ⚙️ Actions
              </h3>
              <div className="space-y-2">
                {deploymentData.status === "pending" && deploymentData.agent_id && (
                  <button
                    onClick={executeDeployment}
                    className="w-full bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
                  >
                    🚀 Exécuter le déploiement
                  </button>
                )}
                {deploymentData.status === "pending" && !deploymentData.agent_id && (
                  <p className="text-sm text-gray-600 italic">
                    Assignez un agent pour exécuter le déploiement
                  </p>
                )}
                {deploymentData.status === "deploying" && (
                  <>
                    <button
                      onClick={() => updateStatus("deployed")}
                      className="w-full bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700"
                    >
                      ✅ Marquer comme déployé
                    </button>
                    <button
                      onClick={() => updateStatus("failed")}
                      className="w-full bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700"
                    >
                      ❌ Marquer comme échoué
                    </button>
                  </>
                )}
                <button
                  onClick={deleteDeployment}
                  className="w-full bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700"
                >
                  🗑️ Supprimer le déploiement
                </button>
              </div>
            </div>
          </div>
        </div>

        {deploymentData.log_output && (
          <div className="border-t pt-6">
            <h3 className="text-xl font-semibold text-gray-800 mb-4">
              📝 Logs de déploiement
            </h3>
            <div className="bg-gray-900 text-green-400 p-4 rounded-lg font-mono text-sm overflow-x-auto">
              <pre className="whitespace-pre-wrap">
                {deploymentData.log_output}
              </pre>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
