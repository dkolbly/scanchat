ScanChat
========

This is a simple REST service for scanning the text of HipChat messages for references
to certain kinds of entities.  Currently supported entities are:

* mentions - A reference to another user.  Starts with an `@` and ends
  with a non-word character ([Reference](https://confluence.atlassian.com/hipchat/get-teammates-attention-744328217.html))

* emoticons - Currently only handles _custom_ emoticons, which are
  alphanumeric strings up to 15 characters in length enclosed in
  parentheses.  ([Reference](https://www.hipchat.com/emoticons))

* links - URLs

To access the service, issue a POST with a Content-Type of `text/plain` to the
server's /scan-message endpoint.

The server returns a JSON document describing the referenced entities.

For example,

    $ curl -d '@donovan you around? (hungry)' http://localhost:8080/scan-message
    {
      "mentions": [
        "donovan"
      ],
      "emoticons": [
        "hungry"
      ]
    }

References to URLs will include the title of the page in the return
value:

    $ curl -d 'Check out this (cat) https://www.youtube.com/watch?v=3EIbWjkimAs' \
                http://localhost:8080/scan-message
    {
      "emoticons": [
        "cat"
      ],
      "links": [
        {
          "url":"https://www.youtube.com/watch?v=3EIbWjkimAs",
          "title":"Funny Cats Compilation (Most Popular) Part 1 - YouTube"
        }
      ]
    }


Security Considerations
-----------------------

This service makes outgoing http(s) connections based on user input.
If not used carefully, it has the potential to expose information from
internal services.  (Note that controlling access by using a public
DNS server is insufficient because an attacker can just configure
their own IN A entry that resolves to a private IP.)

Recommended deployment is to use a firewall rule to disallow outgoing
connections to any private IPs.
