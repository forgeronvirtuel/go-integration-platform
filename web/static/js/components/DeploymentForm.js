function DeploymentForm({ onMessage, onDeploymentCreated }) {
  const [builds, setBuilds] = useState([]);
  const [runners, setRunners] = useState([]);
  const [buildID, setBuildID] = useState("");
  const [runnerID, setRunnerID] = useState("");
  const [loading, setLoading] = useState(false);

  console.log("📋 [DeploymentForm] Component mounted");

  const loadBuilds = async () => {
    try {
      console.log("📋 [DeploymentForm] Loading builds...");
      const data = await API.builds.getAll();
      console.log("📋 [DeploymentForm] Builds loaded:", data);

      const successBuilds = (data || []).filter(
        (b) => b.status && b.status.toLowerCase() === "success"
      );
      console.log("📋 [DeploymentForm] Successful builds:", successBuilds);
      setBuilds(successBuilds);
    } catch (error) {
      console.error("📋 [DeploymentForm] Error loading builds:", error);
    }
  };

  const loadRunners = async () => {
    try {
      const data = await API.runners.getAll("ONLINE");
      setRunners(data.runners || []);
    } catch (error) {
      console.error("📋 [DeploymentForm] Error loading runners:", error);
    }
  };

  React.useEffect(() => {
    loadBuilds();
    loadRunners();
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!buildID) {
      onMessage("Veuillez sélectionner un build");
      return;
    }

    setLoading(true);

    try {
      console.log("📋 [DeploymentForm] Creating deployment...", {
        buildID,
        runnerID,
      });

      const payload = {
        build_id: parseInt(buildID),
      };

      if (runnerID) {
        payload.runner_id = parseInt(runnerID);
      }

      const data = await API.deployments.create(payload);
      console.log("📋 [DeploymentForm] Deployment created:", data);

      onMessage(`Déploiement #${data.id} créé avec succès`);
      onDeploymentCreated();
    } catch (error) {
      console.error("📋 [DeploymentForm] Error creating deployment:", error);
      onMessage(`Erreur: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-white border border-gray-200 rounded-lg p-6">
      <h3 className="text-xl font-semibold mb-4">➕ Nouveau déploiement</h3>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Build à déployer *
          </label>
          <select
            value={buildID}
            onChange={(e) => setBuildID(e.target.value)}
            className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          >
            <option value="">-- Sélectionner un build --</option>
            {builds.map((build) => (
              <option key={build.id} value={build.id}>
                Build #{build.id} - {build.branch} (Projet #{build.project_id})
              </option>
            ))}
          </select>
          {builds.length === 0 && (
            <p className="text-sm text-gray-500 mt-1">
              Aucun build réussi disponible pour le déploiement
            </p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Runner (optionnel)
          </label>
          <select
            value={runnerID}
            onChange={(e) => setRunnerID(e.target.value)}
            className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">-- Assigner plus tard --</option>
            {runners.map((runner) => (
              <option key={runner.id} value={runner.id}>
                {runner.name} (ID: {runner.id})
              </option>
            ))}
          </select>
          {runners.length === 0 && (
            <p className="text-sm text-gray-500 mt-1">
              Aucun runner en ligne disponible
            </p>
          )}
        </div>

        <div className="flex space-x-3 pt-4">
          <button
            type="submit"
            disabled={loading || !buildID}
            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            {loading ? "⏳ Création..." : "🚀 Créer le déploiement"}
          </button>
        </div>
      </form>
    </div>
  );
}
