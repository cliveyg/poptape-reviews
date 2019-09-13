# poptape-reviews
Reviews microservice written in Go

This microservice validates and stores review data in a Postgres database.
Each review is unique per auction and user. i.e. a user cannot leave more
than one review per auction. Editing of a review after creation is not allowed
but reviews can be deleted (though this may change or I may add a deleted flag
if a user wants to remove a review so a user cannot delete a review and add
another).

~~Maybe remove some authentication as all people need to see reviews whether 
authenticated or not.~~

### API routes

```
/reviews [GET] (Authenticated)

Returns a list of reviews for the authenticated user.
Expected normal return codes: [200, 404, 401]


/reviews [POST] (Authenticated)

Create a review for the authenticated user.
Expected normal return codes: [201, 401]


/reviews/<review_id> [GET] (Unauthenticated)

Returns a single review details.
Expected return codes: [200, 404]


/reviews/<review_id> [DELETE] (Authenticated)

Deletes a single review.
Expected return codes: [200, 404]


/reviews/user/<public_id> [GET] (Unauthenticated)

Returns all reviews for a user.
Expected return codes: [200, 404]


/reviews/auction/<auction_id> [GET] (Unauthenticated)

Returns all reviews from a particular auction. As we can have several items
per auction or just one this can vary a lot.
Expected return codes: [200, 404]

```

### To Do:
* ~~Return reviews by auction~~
* ~~Return reviews by user~~
* Need to add check for auction winner
* Write more tests
* Validate input
* Dockerize
* Documentation
