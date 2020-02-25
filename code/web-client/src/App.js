import React, { useState } from 'react';
import { 
  BrowserRouter as Router, 
  Link, 
  Route 
} from 'react-router-dom';
import 'bootstrap/dist/css/bootstrap.min.css';
import './App.css';
import Login from './components/Login';
import Home from './components/Home';
import PrivateRoute from './components/PrivateRoute';
import { AuthContext } from './context/Authentication';

export default function App() {
  // Add state to our App component.
  // Returns the value and a setter.
  const [authTokens, setAuthTokens] = useState(localStorage.getItem('tokens' || undefined));

  const setAuthTokensCallback = (data) => {
    // Callback updates both app state and storage state
    console.log('auth tokens callback called')
    // TODO: try cookies instead of localStorage
    localStorage.setItem('tokens', JSON.stringify(data));
    setAuthTokens(data);
    console.log('new auth tokens: ' + authTokens);
  }

  return (
      <AuthContext.Provider value={{ authTokens, setAuthTokensCallback: setAuthTokensCallback }}>
        <Router>
          <Route path='/login' component={Login} />
          <PrivateRoute exact path='/' component={Home} />
        </Router>
      </AuthContext.Provider>
    );
}
