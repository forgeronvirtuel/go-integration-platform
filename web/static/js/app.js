// Application principale
const { useState } = React;

function App() {
  const [activeTab, setActiveTab] = useState("projects");
  const [message, setMessage] = useState("");

  return (
    <div className="min-h-screen">
      <Header />

      <main className="container mx-auto px-4 py-8">
        <MessageBanner message={message} />

        <div className="card bg-white rounded-lg shadow-lg overflow-hidden">
          <nav className="flex border-b">
            <button
              onClick={() => setActiveTab("projects")}
              className={`tab-button px-6 py-4 font-semibold transition-colors ${
                activeTab === "projects"
                  ? "bg-blue-600 text-white"
                  : "text-gray-600 hover:bg-gray-50"
              }`}
            >
              📁 Projets
            </button>
            <button
              onClick={() => setActiveTab("builds")}
              className={`tab-button px-6 py-4 font-semibold transition-colors ${
                activeTab === "builds"
                  ? "bg-blue-600 text-white"
                  : "text-gray-600 hover:bg-gray-50"
              }`}
            >
              🔨 Builds
            </button>
            <button
              onClick={() => setActiveTab("download")}
              className={`tab-button px-6 py-4 font-semibold transition-colors ${
                activeTab === "download"
                  ? "bg-blue-600 text-white"
                  : "text-gray-600 hover:bg-gray-50"
              }`}
            >
              📥 Téléchargements
            </button>
          </nav>

          <div className="p-8">
            {activeTab === "projects" && <ProjectForm onMessage={setMessage} />}
            {activeTab === "builds" && <BuildForm onMessage={setMessage} />}
            {activeTab === "download" && (
              <DownloadForm onMessage={setMessage} />
            )}
          </div>
        </div>

        <footer className="mt-8 text-center text-gray-500 text-sm">
          <p>
            API disponible sur{" "}
            <code className="bg-white px-2 py-1 rounded">/health</code> et{" "}
            <code className="bg-white px-2 py-1 rounded">/api/*</code>
          </p>
        </footer>
      </main>
    </div>
  );
}

ReactDOM.render(<App />, document.getElementById("root"));
