# poptape-reviews
Reviews microservice written in Go

This microservice validates and stores review data in a Postgres database.
Each review is unique per auction and user. i.e. a user cannot leave more
than one review per auction. Editing of a review after creation is not allowed
but reviews can be deleted (though this may change or I may add a deleted flag
if a user wants to remove a review so a user cannot delete a review and add
another).

Maybe remove some authentication as all people need to see reviews whether 
authenticated or not. 

### API routes

```
/reviews [GET] (Authenticated)

Returns a list of reviews for the authenticated user.
Possible return codes: [200, 404, 401, 500, 502]

/reviews/<review_id> [GET] (Authenticated)

Returns a single review details of for the authenticated user.

```

### To Do:
* ~~Return reviews by auction~~
* ~~Return reviews by user~~
* Write tests
* Validate input
* Dockerize
* Documentation
