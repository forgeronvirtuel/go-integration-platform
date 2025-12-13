// Application principale
const { useState } = React;

function App() {
  const [view, setView] = useState("projects"); // "projects", "project-detail", "build-detail", "agents", "agent-detail"
  const [selectedProject, setSelectedProject] = useState(null);
  const [selectedBuild, setSelectedBuild] = useState(null);
  const [selectedAgent, setSelectedAgent] = useState(null);
  const [message, setMessage] = useState("");

  const handleProjectSelect = (project) => {
    setSelectedProject(project);
    setView("project-detail");
  };

  const handleBuildSelect = (build) => {
    setSelectedBuild(build);
    setView("build-detail");
  };

  const handleAgentSelect = (agent) => {
    setSelectedAgent(agent);
    setView("agent-detail");
  };

  const handleBackToProjects = () => {
    setView("projects");
    setSelectedProject(null);
    setSelectedBuild(null);
  };

  const handleBackToProjectDetail = () => {
    setView("project-detail");
    setSelectedBuild(null);
  };

  const handleBackToAgents = () => {
    setView("agents");
    setSelectedAgent(null);
  };

  const handleNavigate = (targetView) => {
    setView(targetView);
    setSelectedProject(null);
    setSelectedBuild(null);
    setSelectedAgent(null);
  };

  return (
    <div className="min-h-screen">
      <Header currentView={view} onNavigate={handleNavigate} />

      <main className="container mx-auto px-4 py-8">
        <MessageBanner message={message} />

        {view === "projects" && (
          <ProjectsList
            onMessage={setMessage}
            onProjectSelect={handleProjectSelect}
          />
        )}

        {view === "project-detail" && selectedProject && (
          <ProjectDetail
            project={selectedProject}
            onMessage={setMessage}
            onBack={handleBackToProjects}
            onBuildSelect={handleBuildSelect}
          />
        )}

        {view === "build-detail" && selectedBuild && (
          <BuildDetail
            build={selectedBuild}
            project={selectedProject}
            onMessage={setMessage}
            onBack={handleBackToProjectDetail}
          />
        )}

        {view === "agents" && (
          <AgentsList
            onMessage={setMessage}
            onAgentSelect={handleAgentSelect}
          />
        )}

        {view === "agent-detail" && selectedAgent && (
          <AgentDetail
            agent={selectedAgent}
            onMessage={setMessage}
            onBack={handleBackToAgents}
          />
        )}

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

// Attendre que React et tous les composants soient chargés
if (typeof React !== "undefined" && typeof ReactDOM !== "undefined") {
  console.log("✅ [App] React chargé, montage de l'application...");

  // Cacher le loader
  const loader = document.getElementById("loading");
  if (loader) {
    loader.style.display = "none";
  }

  ReactDOM.render(<App />, document.getElementById("root"));
  console.log("✅ [App] Application montée avec succès");
} else {
  console.error("❌ [App] React ou ReactDOM non disponible");
}
