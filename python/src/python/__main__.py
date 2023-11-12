import feedparser
import xml.etree.ElementTree as ET

def extract_rss_urls(opml_file):
    # Parse the OPML file
    tree = ET.parse(opml_file)
    root = tree.getroot()

    # Extract RSS URLs from outline elements
    rss_urls = []
    for outline in root.findall('outline'):
        rss_urls.append(outline.attrib.get('xmlUrl'))

    return rss_urls

# Replace 'opml_file.opml' with the actual path to your OPML file
rss_urls = extract_rss_urls('./feeds.opml.xml')
rss_urls = '\n'.join(rss_urls)

with open("feed.txt", 'w') as f:
    f.write(rss_urls)
