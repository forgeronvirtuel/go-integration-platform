// Composant DeploymentsList
const { useState, useEffect } = React;

function DeploymentsList({ onMessage, onDeploymentSelect }) {
  const [deployments, setDeployments] = useState([]);
  const [builds, setBuilds] = useState({});
  const [projects, setProjects] = useState({});
  const [runners, setRunners] = useState({});
  const [loading, setLoading] = useState(true);
  const [showCreateForm, setShowCreateForm] = useState(false);

  console.log("🚀 [DeploymentsList] Component mounted");

  const loadDeployments = async () => {
    try {
      console.log("🚀 [DeploymentsList] Loading deployments...");
      setLoading(true);

      const data = await DeploymentsAPI.getAll();
      console.log("🚀 [DeploymentsList] Deployments loaded:", data);

      setDeployments(data.deployments || []);

      await loadRelatedData(data.deployments || []);

      setLoading(false);
    } catch (error) {
      console.error("🚀 [DeploymentsList] Error loading deployments:", error);
      onMessage(`Erreur lors du chargement des déploiements: ${error.message}`);
      setLoading(false);
    }
  };

  React.useEffect(() => {
    loadDeployments();
  }, []);

  const loadRelatedData = async (deploymentsList) => {
    try {
      console.log(
        "🚀 [DeploymentsList] Loading related data for deployments..."
      );
      const buildsData = await API.builds.getAll();
      console.log(
        "🚀 [DeploymentsList] Builds loaded for related data:",
        buildsData
      );
      const buildsMap = {};
      (buildsData.builds || []).forEach((build) => {
        buildsMap[build.id] = build;
      });
      setBuilds(buildsMap);

      // Charger tous les projects
      const projectsData = await API.projects.getAll();
      const projectsMap = {};
      (projectsData.projects || []).forEach((project) => {
        projectsMap[project.id] = project;
      });
      setProjects(projectsMap);

      // Charger tous les runners
      const runnersData = await API.runners.getAll();
      const runnersMap = {};
      (runnersData.runners || []).forEach((runner) => {
        runnersMap[runner.id] = runner;
      });
      setRunners(runnersMap);
    } catch (error) {
      console.error("🚀 [DeploymentsList] Error loading related data:", error);
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
        className={`px-3 py-1 rounded-full text-sm font-semibold border ${
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
      dateStyle: "short",
      timeStyle: "medium",
    }).format(date);
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-800">🚀 Déploiements</h2>
          <p className="text-gray-600 mt-1">
            {deployments.length} déploiement
            {deployments.length !== 1 ? "s" : ""}
          </p>
        </div>

        <div className="flex space-x-2">
          <button
            onClick={loadDeployments}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
          >
            🔄 Rafraîchir
          </button>
          <button
            onClick={() => setShowCreateForm(!showCreateForm)}
            className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors"
          >
            {showCreateForm ? "❌ Annuler" : "➕ Nouveau déploiement"}
          </button>
        </div>
      </div>

      {showCreateForm && (
        <DeploymentForm
          onMessage={onMessage}
          onDeploymentCreated={() => {
            setShowCreateForm(false);
            loadDeployments();
          }}
        />
      )}

      {deployments.length === 0 ? (
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
          <p className="text-gray-600">Aucun déploiement trouvé</p>
          <p className="text-sm text-gray-500 mt-2">
            Créez un nouveau déploiement pour démarrer
          </p>
        </div>
      ) : (
        <div className="grid gap-4">
          {deployments.map((deployment) => {
            const build = builds[deployment.build_id];
            const project = build ? projects[build.project_id] : null;
            const runner = deployment.runner_id
              ? runners[deployment.runner_id]
              : null;

            return (
              <div
                key={deployment.id}
                className="bg-white border border-gray-200 rounded-lg p-6 hover:shadow-lg transition-shadow cursor-pointer"
                onClick={() => onDeploymentSelect(deployment)}
              >
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <div className="flex items-center space-x-3 mb-4">
                      <h3 className="text-xl font-semibold text-gray-800">
                        Déploiement #{deployment.id}
                      </h3>
                      {getStatusBadge(deployment.status)}
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center text-sm text-gray-600">
                        <span className="font-medium w-32">📦 Projet:</span>
                        <span className="text-blue-600 font-medium">
                          {project
                            ? project.name
                            : `Build #${deployment.build_id}`}
                        </span>
                      </div>

                      <div className="flex items-center text-sm text-gray-600">
                        <span className="font-medium w-32">🔨 Build:</span>
                        <span>
                          #{deployment.build_id}{" "}
                          {build ? `(${build.branch})` : ""}
                        </span>
                      </div>

                      {runner && (
                        <div className="flex items-center text-sm text-gray-600">
                          <span className="font-medium w-32">🤖 Runner:</span>
                          <span>{runner.name}</span>
                        </div>
                      )}

                      <div className="flex items-center text-sm text-gray-600">
                        <span className="font-medium w-32">📅 Créé:</span>
                        <span>{formatDate(deployment.created_at)}</span>
                      </div>

                      {deployment.ended_at && (
                        <div className="flex items-center text-sm text-gray-600">
                          <span className="font-medium w-32">🏁 Terminé:</span>
                          <span>{formatDate(deployment.ended_at)}</span>
                        </div>
                      )}
                    </div>
                  </div>

                  <div className="ml-4">
                    <button className="text-blue-600 hover:text-blue-800 font-medium">
                      Voir détails →
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
