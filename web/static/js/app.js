// Application principale
const { useState } = React;

function App() {
  const [view, setView] = useState("projects"); // "projects", "project-detail", "build-detail"
  const [selectedProject, setSelectedProject] = useState(null);
  const [selectedBuild, setSelectedBuild] = useState(null);
  const [message, setMessage] = useState("");

  const handleProjectSelect = (project) => {
    setSelectedProject(project);
    setView("project-detail");
  };

  const handleBuildSelect = (build) => {
    setSelectedBuild(build);
    setView("build-detail");
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

  return (
    <div className="min-h-screen">
      <Header />

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
