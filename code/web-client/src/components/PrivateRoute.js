import React from 'react';
import { Route, Redirect } from 'react-router-dom';
import { useAuthContext } from '../context/Authentication';

// PrivateRoute is a decorator around any route that is behind authentication.
export default function PrivateRoute({ component: Component, ...rest }) {
    const { authTokens } = useAuthContext();
    const isAuthenticated = authTokens

    console.log("isAuthenticated: " + isAuthenticated)
    return (
        <Route 
            {...rest}
            // A way to pass components to a component.
            render={(props) => 
                isAuthenticated ?
                // render the requested page
                ( <Component {...props} /> ) :
                // redirect to authentication page
                ( <Redirect to={{
                    pathname: "/login",
                    state: { referer: props.location } 
        }}/> 
                )
            }
        />  
    );
}
