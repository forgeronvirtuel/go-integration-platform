// Composant ProjectsList - Liste tous les projets
function ProjectsList({ onMessage, onProjectSelect }) {
  const [projects, setProjects] = React.useState([]);
  const [loading, setLoading] = React.useState(true);
  const [showForm, setShowForm] = React.useState(false);

  const loadProjects = async () => {
    try {
      console.log("🔍 [ProjectsList] Chargement des projets...");
      const response = await fetch("/api/projects");
      const data = await response.json();
      console.log("📦 [ProjectsList] Données reçues:", data);
      
      if (response.ok) {
        // L'API retourne {projects: [...], count: N}
        const projectsArray = data.projects || [];
        console.log("✅ [ProjectsList] Projets chargés:", projectsArray.length, "projet(s)");
        setProjects(projectsArray);
      } else {
        console.error("❌ [ProjectsList] Erreur HTTP:", response.status, data);
        onMessage("❌ Erreur lors du chargement des projets");
        setProjects([]);
      }
    } catch (error) {
      console.error("❌ [ProjectsList] Erreur réseau:", error);
      onMessage("❌ Erreur réseau: " + error.message);
      setProjects([]);
    } finally {
      setLoading(false);
    }
  };

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

// Composant ProjectForm - Formulaire de création
function ProjectForm({ onMessage, onSuccess }) {
  const [projectName, setProjectName] = React.useState("");
  const [repoUrl, setRepoUrl] = React.useState("");
  const [branch, setBranch] = React.useState("main");
  const [subdir, setSubdir] = React.useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      console.log("📤 [ProjectForm] Création d'un projet:", { name: projectName, repo_url: repoUrl });
      const response = await fetch("/api/projects", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: projectName,
          repo_url: repoUrl,
          branch: branch,
          subdir: subdir || undefined,
        }),
      });
      const data = await response.json();
      console.log("📦 [ProjectForm] Réponse reçue:", data);
      
      if (response.ok) {
        console.log("✅ [ProjectForm] Projet créé avec ID:", data.id);
        onMessage("✅ Projet créé avec succès! ID: " + data.id);
        setProjectName("");
        setRepoUrl("");
        setBranch("main");
        setSubdir("");
        if (onSuccess) onSuccess();
      } else {
        console.error("❌ [ProjectForm] Erreur:", data);
        onMessage("❌ Erreur: " + (data.error || "Erreur inconnue"));
      }
    } catch (error) {
      console.error("❌ [ProjectForm] Erreur réseau:", error);
      onMessage("❌ Erreur réseau: " + error.message);
    }
  };

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6 text-gray-800">
        Créer un nouveau projet
      </h2>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Nom du projet *
          </label>
          <input
            type="text"
            required
            value={projectName}
            onChange={(e) => setProjectName(e.target.value)}
            className="form-input w-full px-4 py-2 border border-gray-300 rounded-lg"
            placeholder="my-awesome-project"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            URL du repository Git *
          </label>
          <input
            type="text"
            required
            value={repoUrl}
            onChange={(e) => setRepoUrl(e.target.value)}
            className="form-input w-full px-4 py-2 border border-gray-300 rounded-lg"
            placeholder="https://github.com/user/repo.git"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Branche
            </label>
            <input
              type="text"
              value={branch}
              onChange={(e) => setBranch(e.target.value)}
              className="form-input w-full px-4 py-2 border border-gray-300 rounded-lg"
              placeholder="main"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Sous-répertoire (optionnel)
            </label>
            <input
              type="text"
              value={subdir}
              onChange={(e) => setSubdir(e.target.value)}
              className="form-input w-full px-4 py-2 border border-gray-300 rounded-lg"
              placeholder="services/api"
            />
          </div>
        </div>

        <button
          type="submit"
          className="btn-primary w-full bg-blue-600 text-white py-3 rounded-lg font-semibold hover:bg-blue-700"
        >
          ➕ Créer le projet
        </button>
      </form>
    </div>
  );
}
