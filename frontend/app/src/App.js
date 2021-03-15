import React from 'react';
import logo from './logo.svg';
import './App.css';

import Root from './containers/Root';

function App() {
  return (
    <div className="App">
      <header className="App-header">
        <h1>{ process.env.REACT_APP_ENV_TITLE }</h1>
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          <Root />
        </p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          Learn React
        </a>
      </header>
    </div>
  );
}

export default App;
