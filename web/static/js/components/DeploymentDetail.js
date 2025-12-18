// Composant DeploymentDetail
const { useState, useEffect } = React;

function DeploymentDetail({ deployment, onMessage, onBack }) {
  const [deploymentData, setDeploymentData] = useState(deployment);
  const [build, setBuild] = useState(null);
  const [project, setProject] = useState(null);
  const [runner, setRunner] = useState(null);
  const [availableRunners, setAvailableRunners] = useState([]);
  const [loading, setLoading] = useState(false);
  const [editingRunner, setEditingRunner] = useState(false);
  const [selectedRunnerID, setSelectedRunnerID] = useState("");

  console.log(
    "🔍 [DeploymentDetail] Component mounted with deployment:",
    deployment
  );

  const loadDeploymentDetails = async () => {
    try {
      console.log("🔍 [DeploymentDetail] Loading deployment details...");
      setLoading(true);

      const deploymentData = await API.deployments.getById(deployment.id);
      console.log(
        "🔍 [DeploymentDetail] Loaded deployment data:",
        deploymentData
      );
      setDeploymentData(deploymentData);

      const buildData = await API.builds.getById(deploymentData.build_id);
      console.log("🔍 [DeploymentDetail] Loaded build data:", buildData);
      setBuild(buildData);

      const projectData = await API.projects.getById(buildData.project_id);
      console.log("🔍 [DeploymentDetail] Loaded project data:", projectData);
      setProject(projectData);

      if (deploymentData.runner_id) {
        const runnerData = await API.runners.getById(deploymentData.runner_id);
        console.log("🔍 [DeploymentDetail] Loaded runner data:", runnerData);
        setRunner(runnerData);
      } else {
        setRunner(null);
      }

      setLoading(false);
    } catch (error) {
      console.error("🔍 [DeploymentDetail] Error loading details:", error);
      onMessage(`Erreur lors du chargement: ${error.message}`);
      setLoading(false);
    }
  };

  const loadAvailableRunners = async () => {
    try {
      const data = await API.runners.getAll("ONLINE");
      setAvailableRunners(data.runners || []);
    } catch (error) {
      console.error("🔍 [DeploymentDetail] Error loading runners:", error);
    }
  };

  const updateStatus = async (newStatus) => {
    try {
      console.log("🔍 [DeploymentDetail] Updating status to:", newStatus);

      const data = await API.deployments.updateStatus(deployment.id, newStatus);
      setDeploymentData(data);
      onMessage(`Statut mis à jour: ${newStatus}`);
    } catch (error) {
      console.error("🔍 [DeploymentDetail] Error updating status:", error);
      onMessage(`Erreur: ${error.message}`);
    }
  };

  const updateRunner = async () => {
    try {
      console.log(
        "🔍 [DeploymentDetail] Updating runner to:",
        selectedRunnerID
      );

      const runnerId = selectedRunnerID ? parseInt(selectedRunnerID) : null;
      const data = await API.deployments.updateRunner(deployment.id, runnerId);

      setDeploymentData(data);
      setEditingRunner(false);
      onMessage("Runner mis à jour avec succès");
      loadDeploymentDetails();
    } catch (error) {
      console.error("🔍 [DeploymentDetail] Error updating runner:", error);
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

      await API.deployments.delete(deployment.id);
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

      await API.deployments.execute(deployment.id);
      onMessage("Déploiement lancé avec succès");
      loadDeploymentDetails(); // Rafraîchir les données
    } catch (error) {
      console.error("🔍 [DeploymentDetail] Error executing deployment:", error);
      onMessage(`Erreur: ${error.message}`);
    }
  };

  useEffect(() => {
    loadDeploymentDetails();
    loadAvailableRunners();
  }, []);

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
                  🤖 Runner
                </h3>
                <button
                  onClick={() => {
                    setEditingRunner(!editingRunner);
                    setSelectedRunnerID(deploymentData.runner_id || "");
                  }}
                  className="text-blue-600 hover:text-blue-800 text-sm font-medium"
                >
                  {editingRunner ? "❌ Annuler" : "✏️ Modifier"}
                </button>
              </div>

              {!editingRunner ? (
                runner ? (
                  <div className="text-gray-800">
                    <p>
                      <strong>Nom:</strong> {runner.name}
                    </p>
                    <p>
                      <strong>ID:</strong> #{runner.id}
                    </p>
                    <p>
                      <strong>Statut:</strong> {runner.status}
                    </p>
                  </div>
                ) : (
                  <p className="text-gray-500">Aucun runner assigné</p>
                )
              ) : (
                <div className="space-y-3">
                  <select
                    value={selectedRunnerID}
                    onChange={(e) => setSelectedRunnerID(e.target.value)}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2"
                  >
                    <option value="">-- Aucun runner --</option>
                    {availableRunners.map((a) => (
                      <option key={a.id} value={a.id}>
                        {a.name} (#{a.id})
                      </option>
                    ))}
                  </select>
                  <button
                    onClick={updateRunner}
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
                {deploymentData.runner_id && (
                  <button
                    onClick={executeDeployment}
                    className="w-full bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
                  >
                    🚀 Exécuter le déploiement
                  </button>
                )}
                {deploymentData.status === "pending" &&
                  !deploymentData.runner_id && (
                    <p className="text-sm text-gray-600 italic">
                      Assignez un runner pour exécuter le déploiement
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
