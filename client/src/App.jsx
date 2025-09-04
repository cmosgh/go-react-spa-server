import './App.css';
import horse from './assets/horse.webp';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import About from './About';

function App() {
  return (
    <Router>
      <div className="App">
        <header className="App-header">
          <img src={horse} className="App-logo" alt="logo" />
          <p>
            I AM HERE TO WORK FOR YOU
          </p>
          <nav>
            <Link to="/">Home</Link> |
            <Link to="/about">About</Link>
          </nav>
          <Routes>
            <Route path="/about" element={<About />} />
            <Route path="/" element={
              <a
                className="App-link"
                href="https://google.com"
                target="_blank"
                rel="noopener noreferrer"
              >
                Go to Google
              </a>
            } />
          </Routes>
        </header>
      </div>
    </Router>
  );
}

export default App;