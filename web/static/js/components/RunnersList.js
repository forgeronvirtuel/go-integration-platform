// Composant RunnersList
const { useState, useEffect } = React;

function RunnersList({ onMessage, onRunnerSelect }) {
  const [runners, setRunners] = useState([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState("all");

  console.log("🤖 [RunnersList] Component mounted");

  const loadRunners = async () => {
    try {
      console.log("🤖 [RunnersList] Loading runners...");
      setLoading(true);

      const data = await API.runners.getAll(
        statusFilter !== "all" ? statusFilter : null
      );
      console.log("🤖 [RunnersList] Runners loaded:", data);

      setRunners(data.runners || []);
      setLoading(false);
    } catch (error) {
      console.error("🤖 [RunnersList] Error loading runners:", error);
      onMessage(`Erreur lors du chargement des runners: ${error.message}`);
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRunners();
    // Rafraîchir toutes les 10 secondes
    const interval = setInterval(loadRunners, 10000);
    return () => clearInterval(interval);
  }, [statusFilter]);

  const getStatusBadge = (status) => {
    const statusColors = {
      ONLINE: "bg-green-100 text-green-800 border-green-300",
      OFFLINE: "bg-gray-100 text-gray-800 border-gray-300",
      DRAINING: "bg-yellow-100 text-yellow-800 border-yellow-300",
    };

    return (
      <span
        className={`px-3 py-1 rounded-full text-sm font-semibold border ${
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
      dateStyle: "short",
      timeStyle: "medium",
    }).format(date);
  };

  const getTimeSince = (dateString) => {
    if (!dateString) return "Jamais";
    const now = new Date();
    const date = new Date(dateString);
    const seconds = Math.floor((now - date) / 1000);

    if (seconds < 60) return `il y a ${seconds}s`;
    if (seconds < 3600) return `il y a ${Math.floor(seconds / 60)}m`;
    if (seconds < 86400) return `il y a ${Math.floor(seconds / 3600)}h`;
    return `il y a ${Math.floor(seconds / 86400)}j`;
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
          <h2 className="text-2xl font-bold text-gray-800">
            🤖 Runners / Runners
          </h2>
          <p className="text-gray-600 mt-1">
            {runners.length} runner{runners.length !== 1 ? "s" : ""} enregistré
            {runners.length !== 1 ? "s" : ""}
          </p>
        </div>

        <div className="flex items-center space-x-2">
          <label className="text-sm font-medium text-gray-700">Statut:</label>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">Tous</option>
            <option value="ONLINE">En ligne</option>
            <option value="OFFLINE">Hors ligne</option>
            <option value="DRAINING">En cours d'arrêt</option>
          </select>

          <button
            onClick={loadRunners}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
          >
            🔄 Rafraîchir
          </button>
        </div>
      </div>

      {runners.length === 0 ? (
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
          <p className="text-gray-600">Aucun runner trouvé</p>
          <p className="text-sm text-gray-500 mt-2">
            Démarrez un runner avec:{" "}
            <code className="bg-gray-200 px-2 py-1 rounded">
              ./bin/gip runner --control-plane http://localhost:3000
            </code>
          </p>
        </div>
      ) : (
        <div className="grid gap-4">
          {runners.map((runner) => (
            <div
              key={runner.id}
              className="bg-white border border-gray-200 rounded-lg p-6 hover:shadow-lg transition-shadow cursor-pointer"
              onClick={() => onRunnerSelect(runner)}
            >
              <div className="flex justify-between items-start">
                <div className="flex-1">
                  <div className="flex items-center space-x-3">
                    <h3 className="text-xl font-semibold text-gray-800">
                      {runner.name}
                    </h3>
                    {getStatusBadge(runner.status)}
                  </div>

                  <div className="mt-4 space-y-2">
                    <div className="flex items-center text-sm text-gray-600">
                      <span className="font-medium w-32">ID:</span>
                      <span>#{runner.id}</span>
                    </div>

                    <div className="flex items-center text-sm text-gray-600">
                      <span className="font-medium w-32">Créé:</span>
                      <span>{formatDate(runner.created_at)}</span>
                    </div>

                    <div className="flex items-center text-sm text-gray-600">
                      <span className="font-medium w-32">Dernier signal:</span>
                      <span
                        className={
                          runner.last_seen_at ? "text-blue-600 font-medium" : ""
                        }
                      >
                        {runner.last_seen_at
                          ? getTimeSince(runner.last_seen_at)
                          : "Jamais"}
                      </span>
                    </div>
                  </div>

                  {runner.labels && Object.keys(runner.labels).length > 0 && (
                    <div className="mt-4">
                      <span className="text-sm font-medium text-gray-700">
                        Labels:
                      </span>
                      <div className="flex flex-wrap gap-2 mt-2">
                        {Object.entries(runner.labels).map(([key, value]) => (
                          <span
                            key={key}
                            className="bg-blue-50 text-blue-700 px-3 py-1 rounded-full text-xs font-medium border border-blue-200"
                          >
                            {key}: {value}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}
                </div>

                <div className="ml-4">
                  <button className="text-blue-600 hover:text-blue-800 font-medium">
                    Voir détails →
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
