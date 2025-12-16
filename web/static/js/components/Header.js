// Composant Header
function Header({ currentView, onNavigate }) {
  return (
    <header className="bg-blue-600 text-white shadow-lg">
      <div className="container mx-auto px-4 py-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold">
              🚀 GIP - Go Integration Platform
            </h1>
            <p className="text-blue-100 mt-2">
              Plateforme de build et déploiement pour projets Go
            </p>
          </div>

          <nav className="flex space-x-4">
            <button
              onClick={() => onNavigate("projects")}
              className={`px-4 py-2 rounded-lg transition-colors ${
                currentView === "projects" ||
                currentView === "project-detail" ||
                currentView === "build-detail"
                  ? "bg-blue-700 font-semibold"
                  : "bg-blue-500 hover:bg-blue-700"
              }`}
            >
              📦 Projets
            </button>
            <button
              onClick={() => onNavigate("deployments")}
              className={`px-4 py-2 rounded-lg transition-colors ${
                currentView === "deployments" ||
                currentView === "deployment-detail"
                  ? "bg-blue-700 font-semibold"
                  : "bg-blue-500 hover:bg-blue-700"
              }`}
            >
              🚀 Déploiements
            </button>
            <button
              onClick={() => onNavigate("runners")}
              className={`px-4 py-2 rounded-lg transition-colors ${
                currentView === "runners" || currentView === "runner-detail"
                  ? "bg-blue-700 font-semibold"
                  : "bg-blue-500 hover:bg-blue-700"
              }`}
            >
              🤖 Runners
            </button>
          </nav>
        </div>
      </div>
    </header>
  );
}
