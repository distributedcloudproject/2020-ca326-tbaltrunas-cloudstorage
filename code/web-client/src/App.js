import React, { useState } from 'react';
import { 
  BrowserRouter as Router, 
  Route 
} from 'react-router-dom';
import Container from 'react-bootstrap/Container';
import 'bootstrap/dist/css/bootstrap.min.css';
import Cookies from 'js-cookie';
import './App.css';
import Header from './components/Header';
import Footer from './components/Footer';
import About from './components/About';
import Login from './components/Login';
import Home from './components/Home';
import PrivateRoute from './components/PrivateRoute';
import { AuthContext } from './context/Authentication';

export default function App() {
  // Add state to our App component.
  // Returns the value and a setter.
  const [isAuthenticated, setIsAuthenticated] = useState(Cookies.get('access_token') !== undefined ? true : false);

  const setIsAuthenticatedCallback = (authed) => {
    // Callback updates both app state and storage state
    console.log('auth tokens callback called')
		
    if (!authed) {
      Cookies.remove('access_token');
    }
    setIsAuthenticated(authed);
  }

  return (
      <AuthContext.Provider value={{ isAuthenticated, setIsAuthenticatedCallback: setIsAuthenticatedCallback }}>
        <Header className='header' />
        <Container className='content'>  
          <Router>
            <Route path='/login' component={Login} />
            <Route path='/about' component={About} />
            <PrivateRoute exact path='/' component={Home} />
          </Router>
        </Container>
        <Footer className='footer' />
      </AuthContext.Provider>
    );
}
