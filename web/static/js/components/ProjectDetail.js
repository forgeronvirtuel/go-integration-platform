// Composant ProjectDetail - Détails d'un projet et ses builds
function ProjectDetail({ project, onMessage, onBack, onBuildSelect }) {
  const [builds, setBuilds] = React.useState([]);
  const [loading, setLoading] = React.useState(true);
  const [showBuildForm, setShowBuildForm] = React.useState(false);
  const [buildBranch, setBuildBranch] = React.useState(project.branch);

  const loadBuilds = async () => {
    try {
      console.log(
        "🔍 [ProjectDetail] Chargement des builds pour projet",
        project.id
      );
      const data = await API.builds.getByProjectId(project.id);
      console.log("📦 [ProjectDetail] Builds reçus:", data);

      const buildsArray = Array.isArray(data) ? data : [];
      console.log(
        "✅ [ProjectDetail] Builds chargés:",
        buildsArray.length,
        "build(s)"
      );
      setBuilds(buildsArray);
    } catch (error) {
      console.error("❌ [ProjectDetail] Erreur:", error);
      onMessage("❌ Erreur: " + error.message);
      setBuilds([]);
    } finally {
      setLoading(false);
    }
  };

  React.useEffect(() => {
    loadBuilds();
  }, [project.id]);

  const handleCreateBuild = async (e) => {
    e.preventDefault();
    try {
      console.log(
        "🔨 [ProjectDetail] Création d'un build pour projet",
        project.id,
        "branche:",
        buildBranch
      );

      const data = await API.builds.create({
        project_id: project.id,
        branch: buildBranch,
      });
      console.log("📦 [ProjectDetail] Réponse build:", data);
      console.log("✅ [ProjectDetail] Build créé avec ID:", data.id);

      onMessage("✅ Build lancé avec succès! ID: " + data.id);
      setShowBuildForm(false);
      loadBuilds();
    } catch (error) {
      console.error("❌ [ProjectDetail] Erreur:", error);
      onMessage("❌ Erreur: " + error.message);
    }
  };

  const handleDeleteProject = async () => {
    if (
      !confirm(
        `Êtes-vous sûr de vouloir supprimer le projet "${project.name}" ?\n\nCette action est irréversible et supprimera également tous les builds associés.`
      )
    ) {
      return;
    }

    try {
      console.log("🗑️ [ProjectDetail] Suppression du projet", project.id);
      await API.projects.delete(project.id);
      console.log("✅ [ProjectDetail] Projet supprimé avec succès");
      onMessage("✅ Projet supprimé avec succès");
      onBack();
    } catch (error) {
      console.error("❌ [ProjectDetail] Erreur lors de la suppression:", error);
      onMessage("❌ Erreur lors de la suppression: " + error.message);
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case "success":
        return "bg-green-100 text-green-800";
      case "failed":
        return "bg-red-100 text-red-800";
      case "building":
        return "bg-yellow-100 text-yellow-800";
      default:
        return "bg-gray-100 text-gray-800";
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

  return (
    <div className="space-y-6">
      {/* En-tête du projet */}
      <div className="card bg-white rounded-lg shadow-lg overflow-hidden">
        <div className="p-6 border-b bg-gradient-to-r from-blue-600 to-blue-700 text-white">
          <button
            onClick={onBack}
            className="mb-4 text-blue-100 hover:text-white flex items-center gap-2"
          >
            ← Retour aux projets
          </button>
          <h2 className="text-3xl font-bold">{project.name}</h2>
          <p className="text-blue-100 mt-2 font-mono text-sm">
            {project.repo_url}
          </p>
          <div className="flex gap-3 mt-3">
            <span className="bg-white/20 text-white px-3 py-1 rounded text-sm">
              🌿 {project.branch}
            </span>
            {project.subdir && (
              <span className="bg-white/20 text-white px-3 py-1 rounded text-sm">
                📂 {project.subdir}
              </span>
            )}
          </div>
        </div>

        {/* Formulaire de build */}
        <div className="p-6 border-b">
          <div className="flex justify-between items-center gap-4">
            <button
              onClick={() => setShowBuildForm(!showBuildForm)}
              className="btn-primary bg-green-600 text-white px-4 py-2 rounded-lg font-semibold hover:bg-green-700"
            >
              {showBuildForm ? "❌ Annuler" : "🔨 Lancer un nouveau build"}
            </button>

            <button
              onClick={handleDeleteProject}
              className="bg-red-600 text-white px-4 py-2 rounded-lg font-semibold hover:bg-red-700 transition-colors"
            >
              🗑️ Supprimer le projet
            </button>
          </div>

          {showBuildForm && (
            <form onSubmit={handleCreateBuild} className="mt-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Branche à builder
                </label>
                <input
                  type="text"
                  value={buildBranch}
                  onChange={(e) => setBuildBranch(e.target.value)}
                  className="form-input w-full px-4 py-2 border border-gray-300 rounded-lg"
                  placeholder="main"
                />
              </div>
              <button
                type="submit"
                className="btn-primary bg-green-600 text-white px-6 py-2 rounded-lg font-semibold hover:bg-green-700"
              >
                🚀 Lancer le build
              </button>
            </form>
          )}
        </div>
      </div>

      {/* Liste des builds */}
      <div className="card bg-white rounded-lg shadow-lg overflow-hidden">
        <div className="p-6 border-b">
          <h3 className="text-2xl font-bold text-gray-800">
            🔨 Historique des builds
          </h3>
        </div>

        <div className="p-6">
          {loading ? (
            <p className="text-center text-gray-600 py-8">
              Chargement des builds...
            </p>
          ) : builds.length === 0 ? (
            <p className="text-center text-gray-500 py-8">
              Aucun build trouvé. Lancez votre premier build !
            </p>
          ) : (
            <div className="space-y-3">
              {builds.map((build) => (
                <div
                  key={build.id}
                  onClick={() => onBuildSelect(build)}
                  className="border rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer hover:bg-gray-50"
                >
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center gap-3">
                        <span className="text-lg font-mono font-semibold text-gray-700">
                          Build #{build.id}
                        </span>
                        <span
                          className={`text-xs px-2 py-1 rounded ${getStatusColor(
                            build.status
                          )}`}
                        >
                          {getStatusIcon(build.status)} {build.status}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 mt-1">
                        🌿 Branche:{" "}
                        <span className="font-mono">{build.branch}</span>
                      </p>
                      <p className="text-xs text-gray-500 mt-1">
                        📅 {new Date(build.created_at).toLocaleString("fr-FR")}
                      </p>
                    </div>
                    <span className="text-gray-400 text-2xl">→</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
