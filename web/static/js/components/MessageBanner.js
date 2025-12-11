// Composant MessageBanner
function MessageBanner({ message }) {
  if (!message) return null;

  return (
    <div className="message-banner mb-6 p-4 bg-white rounded-lg shadow border-l-4 border-blue-500">
      <p className="font-mono text-sm">{message}</p>
    </div>
  );
}
