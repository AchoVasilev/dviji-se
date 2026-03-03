package models

import (
	"encoding/xml"
	"fmt"
	"time"

	"server/internal/domain/posts"
)

// RSS Feed structures
type RSS struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language"`
	LastBuildDate string    `xml:"lastBuildDate"`
	AtomLink      AtomLink  `xml:"atom:link"`
	Items         []RSSItem `xml:"item"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type RSSItem struct {
	Title       string  `xml:"title"`
	Link        string  `xml:"link"`
	Description string  `xml:"description"`
	Author      string  `xml:"author,omitempty"`
	PubDate     string  `xml:"pubDate"`
	GUID        RSSGUID `xml:"guid"`
}

type RSSGUID struct {
	Value       string `xml:",chardata"`
	IsPermaLink bool   `xml:"isPermaLink,attr"`
}

// Sitemap structures
type Sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

type SitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// RSSFromPosts creates an RSS feed from posts
func RSSFromPosts(domainPosts []posts.PostWithAuthor, baseURL, feedURL string) RSS {
	var lastBuildDate string
	if len(domainPosts) > 0 && domainPosts[0].PublishedAt.Valid {
		lastBuildDate = domainPosts[0].PublishedAt.Time.Format(time.RFC1123Z)
	} else {
		lastBuildDate = time.Now().Format(time.RFC1123Z)
	}

	return RSS{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: RSSChannel{
			Title:         "Движи се - Фитнес блог",
			Link:          baseURL,
			Description:   "Фитнес съвети, тренировки и здравословен начин на живот",
			Language:      "bg",
			LastBuildDate: lastBuildDate,
			AtomLink: AtomLink{
				Href: feedURL,
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: rssItemsFromPosts(domainPosts, baseURL),
		},
	}
}

func rssItemsFromPosts(domainPosts []posts.PostWithAuthor, baseURL string) []RSSItem {
	items := make([]RSSItem, 0, len(domainPosts))

	for _, post := range domainPosts {
		var pubDate string
		if post.PublishedAt.Valid {
			pubDate = post.PublishedAt.Time.Format(time.RFC1123Z)
		}

		description := post.Excerpt
		if description == "" {
			description = post.MetaDescription
		}

		postURL := fmt.Sprintf("%s/blog/%s", baseURL, post.Slug)

		items = append(items, RSSItem{
			Title:       post.Title,
			Link:        postURL,
			Description: description,
			Author:      fmt.Sprintf("%s %s", post.AuthorFirstName, post.AuthorLastName),
			PubDate:     pubDate,
			GUID: RSSGUID{
				Value:       postURL,
				IsPermaLink: true,
			},
		})
	}

	return items
}

// SitemapFromPosts creates a sitemap from posts
func SitemapFromPosts(domainPosts []posts.PostWithAuthor, baseURL string) Sitemap {
	urls := []SitemapURL{
		{
			Loc:        baseURL,
			ChangeFreq: "daily",
			Priority:   "1.0",
		},
		{
			Loc:        baseURL + "/blog",
			ChangeFreq: "daily",
			Priority:   "0.9",
		},
	}

	for _, post := range domainPosts {
		var lastMod string
		if post.UpdatedAt.Valid {
			lastMod = post.UpdatedAt.Time.Format("2006-01-02")
		} else if post.PublishedAt.Valid {
			lastMod = post.PublishedAt.Time.Format("2006-01-02")
		}

		urls = append(urls, SitemapURL{
			Loc:        fmt.Sprintf("%s/blog/%s", baseURL, post.Slug),
			LastMod:    lastMod,
			ChangeFreq: "weekly",
			Priority:   "0.8",
		})
	}

	return Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}
}
