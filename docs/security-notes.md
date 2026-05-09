# Security Notes

## Authentication vs Authorization
### Authentication
It is used for checking where a person is actual User of the application or not.

This is done in few ways,
- email and password
- email and security code
- phone number and security code
- third party signin using gmail, github login
- security devices
- finger prints
- facial recognition

Also, multiple steps are kept depending on how secure the application need to be and this is called as MFA (Multi Factor Authentication)

### Authorization
It is used check if a authenticated user, has necessary access to use a particular feature or data or device within the system.

This is done few ways,
- role based, the user has, admin, user, superadmin, ect. Each user can have one or more roles.
- permission based, which are delete, read, create, update, only admin can do all operation, other user will have ony read and create.
- attribute based, to restrict access b/w multiple departments, lab department can't be accessed by other department. 
- ownership based, one the person who posted the photo can delete it.
- multi-tenant based, adding tenent added for each user, when same application will be used by multiple companies, so user with this tenant id belong to this company.
- scope based, check if user has scope for doing a particular action, scope:"read:profile read:orders" then I can't do write on either profile or orders.
- policy based, we can have a centralised policy that will be followed in the application. For example, user can update order if they have admin, order owner or assigned to that  region, used in complex enterprise applications.
- relationship based, if user A owns project X, then he has full access to it. Mostly used on data and file sharing applications.
- feature based, check if user has access to this particular feature or not, used for SASS application subscription model.
- database level, in database we can enforce policy that checks if a query is done by the user, is there tenent_id belong to the company data they are accessing. used in high level security companies.  

## Access Token vs Refresh Token

Access token is temporary entry pass
Refresh token is renewal pass.

{
  "accessToken": "short-lived-token",
  "refreshToken": "long-lived-token"
}

## Why Refresh Token Rotation Matters
Refresh token are long lived and it might get stolen, then someone can use it get access token.
Due to rotation, a refresh token can be used only once.


## Common Auth Risks
- weak password
- storage of token in wrong place
- long-lived access token
- no refresh token rotation
- no logout in backend means not deleting token on backend.
- no rate limiter
- account enumuration meaning that show invalid password or email and not invalid email, it helps hacker find out valid email.
- insecure reset
- weak jwt validation
- CSRF and XSS attack
- poor error handling
- bad MFA
- poor CORS configuration
- Not using Https
- No audit

## Current Security Features in This Project
- using firebaseSDK
- JWT token
- short lived access token
- long lived refresh token
- refresh token rotate
- proper cookie config
- prevented CSRF and XSS attacks
- rate limiter using in-memory