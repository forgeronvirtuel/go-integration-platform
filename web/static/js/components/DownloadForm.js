// Composant DownloadForm
function DownloadForm({ onMessage }) {
  const [downloadBuildId, setDownloadBuildId] = React.useState("");

  const handleDownload = () => {
    if (!downloadBuildId) {
      onMessage("❌ Veuillez entrer un Build ID");
      return;
    }
    window.location.href = "/v1/api/builds/" + downloadBuildId + "/download";
    onMessage("📥 Téléchargement lancé...");
  };

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6 text-gray-800">
        Télécharger un binaire
      </h2>
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            ID du build *
          </label>
          <input
            type="number"
            value={downloadBuildId}
            onChange={(e) => setDownloadBuildId(e.target.value)}
            className="form-input w-full px-4 py-2 border border-gray-300 rounded-lg"
            placeholder="1"
          />
        </div>

        <button
          onClick={handleDownload}
          className="btn-primary w-full bg-purple-600 text-white py-3 rounded-lg font-semibold hover:bg-purple-700"
        >
          📥 Télécharger le binaire
        </button>
      </div>

      <div className="mt-6 p-4 bg-yellow-50 rounded-lg border border-yellow-200">
        <p className="text-sm text-gray-700">
          <strong>⚠️ Note:</strong> Seuls les builds avec le statut "success"
          peuvent être téléchargés. Le binaire est un exécutable Go compilé.
        </p>
      </div>
    </div>
  );
}
