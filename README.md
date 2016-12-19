ScanChat
========

Requires Go 1.7+.  Installing:

    go get github.com/dkolbly/scanchat

By default it listens on port 8080, but this can be changed with the -http flag:

    $GOBIN/scanchat -http 8111

This is a simple REST service for scanning the text of HipChat
messages for references to certain kinds of entities.  Currently
supported entities are:

* mentions - A reference to another user.  Starts with an `@` and ends
  with a non-word character
  ([Reference](https://confluence.atlassian.com/hipchat/get-teammates-attention-744328217.html))

* emoticons - Currently only handles _custom_ emoticons, which are
  alphanumeric strings up to 15 characters in length enclosed in
  parentheses.  ([Reference](https://www.hipchat.com/emoticons))

* links - URLs.  Must start with `http:` or `https:` and are parsed
  somewhat inexactly but with a best effort at working well in the
  context of URLs in chat messages.

To use the service, issue a POST with a Content-Type of `text/plain`
to the server's `/scan-message` endpoint and with a request entity
containing the message to be scanned.

> ⚠ **NOTE** This service is not designed to handle arbitrarily large
> inputs; it will buffer the input for ease of processing, which is
> suitable for the expected inputs up to the handful of MB range.

The server returns a JSON document describing the referenced entities.

For example,

    $ curl --data-raw '@donovan you around? (hungry)' http://localhost:8080/scan-message
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

For URL links, the returned title is *not* HTML-escaped, so if it is
to be interpolated into a web page (_e.g._, if implementing a web
client using this service), it *must* be done using safe means
appropriate to the JS framework being used.

    $ curl -d 'Not HTML encoded (lol) https://processing.org/reference/lessthan.html' \
           http://localhost:8080/scan-message
    {
      "emoticons": [
        "lol"
      ],
      "links": [
        {
          "url": "https://processing.org/reference/lessthan.html",
          "title": "\u003c (less than) \\ Language (API) \\ Processing 2+"
        }
      ]
    }

If there is an error attempting to dereference a link, it will be returned
without a `title` property but with an `error` property:

    $ curl -d 'No such thing http://www.rscheme.org/no-such-thing' \
                http://localhost:8080/scan-message
    {
      "links": [
        {
          "url": "http://www.rscheme.org/no-such-thing",
          "error": "404 Not Found"
        }
      ]
    }

    $ curl -d 'No such thing http://no-such-host.org/' \
                http://localhost:8080/scan-message
    {
      "links": [
        {
          "url": "http://no-such-host.org/",
          "error": "Get http://no-such-host.org/: dial tcp: lookup no-such-host.org on 192.168.1.1:53: no such host"
        }
      ]
    }


External Libraries
------------------

This service makes use of
[golang.org/x/net/html](https://godoc.org/golang.org/x/net/html) for
parsing the HTML documents returned during link dereferencing.  This library
seems to be fairly robust at handling HTML pages seen in the wild, at
least for parsing the &lt;title&gt; element, which is all we need.

Security Considerations
-----------------------

This service makes outgoing http(s) connections based on user input.
If not used carefully, it has the potential to expose information from
internal services.  (Note that controlling access by using a public
DNS server is insufficient because an attacker can just configure
their own IN A entry that resolves to a private IP.)

Recommended deployment is to use a firewall rule to disallow outgoing
connections to any private IPs.

Caching Considerations
----------------------

Because this service accesses arbitrary pages on the internet, it
might be advisable to deploy a proxy cache such as
[squid](http://www.squid-cache.org/) to take advantage of temporal
locality in page references.  There are containerized implementations
available for squid, e.g.,
[docker-squid](https://github.com/sameersbn/docker-squid) for easy
deployment.

In this case, set the `HTTP_PROXY` environment variable to the host
and port (_e.g._, `localhost:3128`) when running this service, _e.g._:

    docker run --name squid -d -p 3128:3128 sameersbn/squid:3.3.8-20
    export HTTP_PROXY=localhost:3128
    $GOBIN/scanchat

It would also be advisable to collect operational statistics to
determine the potential efficacy of different caching strategies.  For
example, if many links are `https` or most pages access have no-cache
headers set, a simple proxy may not do much.

Furthermore, application level caching may be even more effective, and
should be investigated for future work.  In particular:

* it would not suffer from https opacity,

* it could force a minimal level of cache retention in spite of HTTP
  header settings, and

* it can save even the work of re-parsing cached HTML, since all we
  need is the value of the &lt;title&gt; element.

For application-level caching of titles,
[memcache](https://memcached.org/) or [redis](https://redis.io/) would
both probably work pretty well because expiration times can be set on
a per-entry basis to respect the HTTP headers (there are other
solutions as well, no doubt, but something like `groupcache`, while
being nice for scaling out and avoiding the thundering herd problem,
doesn't support expiration.)
