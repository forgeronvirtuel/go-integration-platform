// Application principale
const { useState } = React;

function App() {
  const [view, setView] = useState("projects"); // "projects", "project-detail", "build-detail", "runners", "runner-detail", "deployments", "deployment-detail"
  const [selectedProject, setSelectedProject] = useState(null);
  const [selectedBuild, setSelectedBuild] = useState(null);
  const [selectedRunner, setSelectedRunner] = useState(null);
  const [selectedDeployment, setSelectedDeployment] = useState(null);
  const [message, setMessage] = useState("");

  const handleProjectSelect = (project) => {
    setSelectedProject(project);
    setView("project-detail");
  };

  const handleBuildSelect = (build) => {
    setSelectedBuild(build);
    setView("build-detail");
  };

  const handleRunnerSelect = (runner) => {
    setSelectedRunner(runner);
    setView("runner-detail");
  };

  const handleDeploymentSelect = (deployment) => {
    setSelectedDeployment(deployment);
    setView("deployment-detail");
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

  const handleBackToRunners = () => {
    setView("runners");
    setSelectedRunner(null);
  };

  const handleBackToDeployments = () => {
    setView("deployments");
    setSelectedDeployment(null);
  };

  const handleNavigate = (targetView) => {
    setView(targetView);
    setSelectedProject(null);
    setSelectedBuild(null);
    setSelectedRunner(null);
    setSelectedDeployment(null);
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

        {view === "runners" && (
          <RunnersList
            onMessage={setMessage}
            onRunnerSelect={handleRunnerSelect}
          />
        )}

        {view === "runner-detail" && selectedRunner && (
          <RunnerDetail
            runner={selectedRunner}
            onMessage={setMessage}
            onBack={handleBackToRunners}
          />
        )}

        {view === "deployments" && (
          <DeploymentsList
            onMessage={setMessage}
            onDeploymentSelect={handleDeploymentSelect}
          />
        )}

        {view === "deployment-detail" && selectedDeployment && (
          <DeploymentDetail
            deployment={selectedDeployment}
            onMessage={setMessage}
            onBack={handleBackToDeployments}
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
