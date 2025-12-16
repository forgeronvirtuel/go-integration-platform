// Composant ProjectsList - Liste tous les projets
function ProjectsList({ onMessage, onProjectSelect }) {
  const [projects, setProjects] = React.useState([]);
  const [loading, setLoading] = React.useState(true);
  const [showForm, setShowForm] = React.useState(false);

  const loadProjects = async () => {
    try {
      console.log("🔍 [ProjectsList] Chargement des projets...");
      const data = await API.projects.getAll();
      console.log("📦 [ProjectsList] Données reçues:", data);

      // L'API retourne {projects: [...], count: N}
      const projectsArray = data.projects || [];
      console.log(
        "✅ [ProjectsList] Projets chargés:",
        projectsArray.length,
        "projet(s)"
      );
      setProjects(projectsArray);
    } catch (error) {
      console.error("❌ [ProjectsList] Erreur réseau:", error);
      onMessage("❌ Erreur réseau: " + error.message);
      setProjects([]);
    } finally {
      setLoading(false);
    }
  };

  //   try {
  //     console.log("🔍 [ProjectsList] Chargement des projets...");
  //     const response = await ProjectsAPI.getAll();
  //     console.log("📦 [ProjectsList] Données reçues:", data);

  //     if (response.ok) {
  //       // L'API retourne {projects: [...], count: N}
  //       const projectsArray = data.projects || [];
  //       console.log(
  //         "✅ [ProjectsList] Projets chargés:",
  //         projectsArray.length,
  //         "projet(s)"
  //       );
  //       setProjects(projectsArray);
  //     } else {
  //       console.error("❌ [ProjectsList] Erreur HTTP:", response.status, data);
  //       onMessage("❌ Erreur lors du chargement des projets");
  //       setProjects([]);
  //     }
  //   } catch (error) {
  //     console.error("❌ [ProjectsList] Erreur réseau:", error);
  //     onMessage("❌ Erreur réseau: " + error.message);
  //     setProjects([]);
  //   } finally {
  //     setLoading(false);
  //   }
  // };

  React.useEffect(() => {
    loadProjects();
  }, []);

  const handleProjectCreated = () => {
    setShowForm(false);
    loadProjects();
  };

  if (loading) {
    return (
      <div className="card bg-white rounded-lg shadow-lg p-8">
        <p className="text-center text-gray-600">Chargement des projets...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="card bg-white rounded-lg shadow-lg overflow-hidden">
        <div className="p-6 border-b flex justify-between items-center">
          <h2 className="text-2xl font-bold text-gray-800">📁 Mes Projets</h2>
          <button
            onClick={() => setShowForm(!showForm)}
            className="btn-primary bg-blue-600 text-white px-4 py-2 rounded-lg font-semibold hover:bg-blue-700"
          >
            {showForm ? "❌ Annuler" : "➕ Nouveau Projet"}
          </button>
        </div>

        {showForm && (
          <div className="p-6 border-b bg-gray-50">
            <ProjectForm
              onMessage={onMessage}
              onSuccess={handleProjectCreated}
            />
          </div>
        )}

        <div className="p-6">
          {projects.length === 0 ? (
            <p className="text-center text-gray-500 py-8">
              Aucun projet trouvé. Créez votre premier projet !
            </p>
          ) : (
            <div className="grid gap-4">
              {projects.map((project) => (
                <div
                  key={project.id}
                  onClick={() => onProjectSelect(project)}
                  className="border rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer hover:bg-gray-50"
                >
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <h3 className="text-lg font-semibold text-gray-800">
                        {project.name}
                      </h3>
                      <p className="text-sm text-gray-600 mt-1 font-mono truncate">
                        {project.repo_url}
                      </p>
                      <div className="flex gap-3 mt-2">
                        <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                          🌿 {project.branch}
                        </span>
                        {project.subdir && (
                          <span className="text-xs bg-purple-100 text-purple-800 px-2 py-1 rounded">
                            📂 {project.subdir}
                          </span>
                        )}
                      </div>
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
