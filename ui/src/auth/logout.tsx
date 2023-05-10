import React from "react";
import { Navigate } from "react-router-dom";
import Meter from "../info/meter";
import { urlAuthError } from "./auth_error";

enum logoutNav {
  waiting,
  completed,
  error  
}

interface LogoutState{
  logoutNav: logoutNav;
}

interface URLProps {
  redirectURL: string;
}

// Logout logs out and redirects to the page sent as a parameter
export default class Logout extends React.Component<URLProps, LogoutState> {

  constructor(props: URLProps) {
    super(props);

    this.state = {
      logoutNav: logoutNav.waiting,
    };
  }

  async componentDidMount() {
    try {
      const res = await fetch("/auth/api/logout", {method: "GET"});
      if (res.status == 200) {
        this.setState({logoutNav: logoutNav.completed});
        return
      }
      console.log("auth.logout. Fetch to logout server. Status: " + res.status);
      this.setState({logoutNav: logoutNav.error});
    } catch (ex) {
        console.log("auth.logout. Fetch to logout server: " + ex);
        this.setState({logoutNav: logoutNav.error});
    }
  }

  render(): React.ReactNode {
    switch (this.state.logoutNav) {
      case logoutNav.completed:
        return <Navigate to={this.props.redirectURL} replace={true} />;
        break;
      case logoutNav.error:
        return <Navigate to={urlAuthError} replace={true} />;
        break;
      default:
        return <Meter message="Redirigiendo a la paǵina de logout..."/>;
        break;
    }
  }
}
