// Main App component

import React from 'react';
import { ChatPanel } from './components';

function App() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-100 to-gray-200 p-4 md:p-8">
      <div className="max-w-4xl mx-auto h-[calc(100vh-4rem)]">
        <ChatPanel domain="cio" />
      </div>
    </div>
  );
}

export default App;
