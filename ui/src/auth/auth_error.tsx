import Ups from "../info/ups";

export const urlAuthError = "/auth/error"

// Error reports access errors
export default function AuthError() {
    return (
        <Ups message="Se ha producido un problema en la autenticación. Si persite el problema póngase en contacto con el adminstrador"/>
      );    
}