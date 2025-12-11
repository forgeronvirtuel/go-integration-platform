// Composant ProjectForm
function ProjectForm({ onMessage }) {
  const [projectName, setProjectName] = React.useState("");
  const [repoUrl, setRepoUrl] = React.useState("");
  const [branch, setBranch] = React.useState("main");
  const [subdir, setSubdir] = React.useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
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
      if (response.ok) {
        onMessage("✅ Projet créé avec succès! ID: " + data.id);
        setProjectName("");
        setRepoUrl("");
        setBranch("main");
        setSubdir("");
      } else {
        onMessage("❌ Erreur: " + (data.error || "Erreur inconnue"));
      }
    } catch (error) {
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
