// Composant BuildDetail - Détails d'un build
function BuildDetail({ build, project, onMessage, onBack }) {
  const [buildData, setBuildData] = React.useState(build);
  const [loading, setLoading] = React.useState(false);

  const refreshBuild = async () => {
    setLoading(true);
    try {
      console.log("🔄 [BuildDetail] Rafraîchissement du build", build.id);
      const response = await fetch(`/v1/api/builds/${build.id}`);
      const data = await response.json();
      console.log("📦 [BuildDetail] Données du build:", data);

      if (response.ok) {
        console.log("✅ [BuildDetail] Build mis à jour, statut:", data.status);
        setBuildData(data);
      } else {
        console.error("❌ [BuildDetail] Erreur HTTP:", response.status, data);
        onMessage("❌ Erreur lors du rafraîchissement");
      }
    } catch (error) {
      console.error("❌ [BuildDetail] Erreur réseau:", error);
      onMessage("❌ Erreur réseau: " + error.message);
    } finally {
      setLoading(false);
    }
  };

  React.useEffect(() => {
    // Auto-refresh si le build est en cours
    if (buildData.status === "building" || buildData.status === "pending") {
      const interval = setInterval(refreshBuild, 3000);
      return () => clearInterval(interval);
    }
  }, [buildData.status]);

  const getStatusColor = (status) => {
    switch (status) {
      case "success":
        return "bg-green-100 text-green-800 border-green-300";
      case "failed":
        return "bg-red-100 text-red-800 border-red-300";
      case "building":
        return "bg-yellow-100 text-yellow-800 border-yellow-300";
      default:
        return "bg-gray-100 text-gray-800 border-gray-300";
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case "success":
        return "✅";
      case "failed":
        return "❌";
      case "building":
        return "⚙️";
      default:
        return "⏳";
    }
  };

  const handleDownload = () => {
    window.location.href = `/v1/api/builds/${buildData.id}/download`;
    onMessage("📥 Téléchargement lancé...");
  };

  return (
    <div className="space-y-6">
      {/* En-tête */}
      <div className="card bg-white rounded-lg shadow-lg overflow-hidden">
        <div className="p-6 border-b bg-gradient-to-r from-gray-700 to-gray-800 text-white">
          <button
            onClick={onBack}
            className="mb-4 text-gray-300 hover:text-white flex items-center gap-2"
          >
            ← Retour au projet
          </button>
          <div className="flex justify-between items-start">
            <div>
              <h2 className="text-3xl font-bold">Build #{buildData.id}</h2>
              <p className="text-gray-300 mt-2">
                Projet: <span className="font-semibold">{project.name}</span>
              </p>
            </div>
            <button
              onClick={refreshBuild}
              disabled={loading}
              className="bg-white/20 hover:bg-white/30 text-white px-4 py-2 rounded-lg"
            >
              🔄 Rafraîchir
            </button>
          </div>
        </div>
      </div>

      {/* Informations du build */}
      <div className="card bg-white rounded-lg shadow-lg overflow-hidden">
        <div className="p-6 border-b">
          <h3 className="text-xl font-bold text-gray-800">📊 Informations</h3>
        </div>
        <div className="p-6 space-y-4">
          <div className="grid grid-cols-2 gap-6">
            <div>
              <p className="text-sm text-gray-600 mb-1">Statut</p>
              <span
                className={`inline-block px-4 py-2 rounded-lg border-2 font-semibold ${getStatusColor(
                  buildData.status
                )}`}
              >
                {getStatusIcon(buildData.status)}{" "}
                {buildData.status.toUpperCase()}
              </span>
            </div>
            <div>
              <p className="text-sm text-gray-600 mb-1">Branche</p>
              <span className="inline-block bg-blue-100 text-blue-800 px-4 py-2 rounded-lg font-mono">
                🌿 {buildData.branch}
              </span>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-6">
            <div>
              <p className="text-sm text-gray-600 mb-1">Créé le</p>
              <p className="font-mono text-gray-800">
                📅 {new Date(buildData.created_at).toLocaleString("fr-FR")}
              </p>
            </div>
            {buildData.started_at && (
              <div>
                <p className="text-sm text-gray-600 mb-1">Démarré le</p>
                <p className="font-mono text-gray-800">
                  ⏱️ {new Date(buildData.started_at).toLocaleString("fr-FR")}
                </p>
              </div>
            )}
          </div>

          {buildData.ended_at && (
            <div>
              <p className="text-sm text-gray-600 mb-1">Terminé le</p>
              <p className="font-mono text-gray-800">
                🏁 {new Date(buildData.ended_at).toLocaleString("fr-FR")}
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Actions */}
      {buildData.status === "success" && (
        <div className="card bg-white rounded-lg shadow-lg overflow-hidden">
          <div className="p-6 border-b bg-green-50">
            <h3 className="text-xl font-bold text-gray-800">
              📥 Téléchargement
            </h3>
          </div>
          <div className="p-6">
            <button
              onClick={handleDownload}
              className="btn-primary w-full bg-purple-600 text-white py-4 rounded-lg font-semibold hover:bg-purple-700 text-lg"
            >
              📥 Télécharger le binaire
            </button>
            <p className="text-sm text-gray-600 mt-3 text-center">
              Le binaire compilé sera téléchargé sur votre machine
            </p>
          </div>
        </div>
      )}

      {/* Logs */}
      {buildData.log_output && (
        <div className="card bg-white rounded-lg shadow-lg overflow-hidden">
          <div className="p-6 border-b bg-gray-800 text-white">
            <h3 className="text-xl font-bold">📜 Logs du build</h3>
          </div>
          <div className="p-6 bg-gray-900">
            <pre className="text-xs text-green-400 font-mono overflow-x-auto whitespace-pre-wrap">
              {buildData.log_output}
            </pre>
          </div>
        </div>
      )}

      {buildData.status === "building" && (
        <div className="card bg-white rounded-lg shadow-lg overflow-hidden">
          <div className="p-6 text-center">
            <div className="animate-spin inline-block w-12 h-12 border-4 border-yellow-400 border-t-transparent rounded-full mb-4"></div>
            <p className="text-gray-600 font-semibold">Build en cours...</p>
            <p className="text-sm text-gray-500 mt-2">
              Cette page se rafraîchit automatiquement
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
