// Composant DeploymentForm
const { useState, useEffect } = React;

function DeploymentForm({ onMessage, onDeploymentCreated }) {
  const [builds, setBuilds] = useState([]);
  const [agents, setAgents] = useState([]);
  const [buildID, setBuildID] = useState("");
  const [agentID, setAgentID] = useState("");
  const [loading, setLoading] = useState(false);

  console.log("📋 [DeploymentForm] Component mounted");

  useEffect(() => {
    loadBuilds();
    loadAgents();
  }, []);

  const loadBuilds = async () => {
    try {
      const response = await fetch("/v1/builds");
      const data = await response.json();
      // Filtrer uniquement les builds réussis
      const successBuilds = (data.builds || []).filter(
        (b) => b.status === "success"
      );
      setBuilds(successBuilds);
    } catch (error) {
      console.error("📋 [DeploymentForm] Error loading builds:", error);
    }
  };

  const loadAgents = async () => {
    try {
      const response = await fetch("/v1/api/agents?status=ONLINE");
      const data = await response.json();
      setAgents(data.agents || []);
    } catch (error) {
      console.error("📋 [DeploymentForm] Error loading agents:", error);
    }
  };

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
        agentID,
      });

      const payload = {
        build_id: parseInt(buildID),
      };

      if (agentID) {
        payload.agent_id = parseInt(agentID);
      }

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

      const data = await response.json();
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
            Agent (optionnel)
          </label>
          <select
            value={agentID}
            onChange={(e) => setAgentID(e.target.value)}
            className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">-- Assigner plus tard --</option>
            {agents.map((agent) => (
              <option key={agent.id} value={agent.id}>
                {agent.name} (ID: {agent.id})
              </option>
            ))}
          </select>
          {agents.length === 0 && (
            <p className="text-sm text-gray-500 mt-1">
              Aucun agent en ligne disponible
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
