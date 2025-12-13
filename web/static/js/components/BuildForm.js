// Composant BuildForm
function BuildForm({ onMessage }) {
  const [buildProjectId, setBuildProjectId] = React.useState("");
  const [buildBranch, setBuildBranch] = React.useState("main");

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch("/v1/api/builds/", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          project_id: parseInt(buildProjectId),
          branch: buildBranch,
        }),
      });
      const data = await response.json();
      if (response.ok) {
        onMessage(
          "✅ Build lancé avec succès! ID: " +
            data.id +
            " - Status: " +
            data.status
        );
        setBuildProjectId("");
        setBuildBranch("main");
      } else {
        onMessage("❌ Erreur: " + (data.error || "Erreur inconnue"));
      }
    } catch (error) {
      onMessage("❌ Erreur réseau: " + error.message);
    }
  };

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6 text-gray-800">Lancer un build</h2>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            ID du projet *
          </label>
          <input
            type="number"
            required
            value={buildProjectId}
            onChange={(e) => setBuildProjectId(e.target.value)}
            className="form-input w-full px-4 py-2 border border-gray-300 rounded-lg"
            placeholder="1"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Branche
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
          className="btn-primary w-full bg-green-600 text-white py-3 rounded-lg font-semibold hover:bg-green-700"
        >
          🔨 Lancer le build
        </button>
      </form>

      <div className="mt-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
        <p className="text-sm text-gray-700">
          <strong>💡 Astuce:</strong> Le build clone le repository, exécute{" "}
          <code className="bg-white px-2 py-1 rounded">go mod download</code> et
          compile le binaire avec{" "}
          <code className="bg-white px-2 py-1 rounded">
            go build ./cmd/main.go
          </code>
        </p>
      </div>
    </div>
  );
}
