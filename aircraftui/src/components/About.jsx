import React from 'react';
import Tabs from 'react-bootstrap/Tabs';
import Tab from 'react-bootstrap/Tab';

class About extends React.Component {
  render() {
    return (
      <Tabs defaultActiveKey='about-app'>
        <Tab eventKey='about-app' title="Uses &amp; Data">
          <p>The Live Situation application provides a single holistic
            view of data flowing through the VMware Tanzu Platform. It is a Situational Awareness application that provides a single pane of glass to an organisation's users about their geospatial data.
          </p>
          <p>
            The core of the platform is a messaging layer than can receive any data. Multiple data processors can be installed and configured in order to provide live and historic views of objects and their tracks.
          </p>
          <p>
            A unified user interface provides a consistent look and feel no matter what type of source data is used in the platform. This can be easily customised to add organisation specific annotations to the data, providing a familiar view to end users.
          </p>
        </Tab>
        <Tab eventKey='about-arch' title="Design &amp; Architecture">
          <p>This application consists of a web user interface at the front that can display a wide range of ground, sea and air vehicles, and a set of supporting micro services behind it.</p>

          <img src="/about/adsb-architecture-full.png" alt="Architecture diagram" />

          <ol>
            <li>Edge deployed Internet of Things (IoT) devices receive transmissions sent on 1090 MHz (ADS-B aircraft transmissions)</li>
            <li>RabbitMQ receives these messages and sends copies to each type of processor (live and historic)</li>
            <li>A message driven micro service, the live processor, takes each message and updates the current position of each aircraft in Redis. Each record has a time to live (TTL) of 60 seconds. If the data reaches 60 seconds in age, it is removed automatically from the database. This provides a 'live view'</li>
            <li>A RESTful microservice, aircraft-monitor, receives requests from an end user and fetches the current 'live view' of all aircraft</li>
            <li>A ReactJS based, fully client side, web application is downloaded by the user's web browser. This requests all information from the RESTful micro service</li>
          </ol>
        </Tab>
        <Tab eventKey='about-software' title="Software used">
          <p>There are various application platform, data platform, and end user software products used in this application.</p>
          <ul>
            <li>The <a href="https://tanzu.vmware.com" target="_pio">VMware Tanzu Platform</a> is used to host the micro services, web application, and data layers of the application. This provides a consistent platform no matter the infrastructure (private or public cloud) used</li>
            <li><a href="https://www.rabbitmq.com/" target="_rabbitmq">VMware Tanzu RabbitMQ</a> provides the reliable messaging layer where data is received, transformed, and fused</li>
            <li><a href="https://redis.io" target="_redis">Opensource Redis</a> is used as a 'live view' cache</li>
            <li><a href="https://tanzu.vmware.com/greenplum" target="_greenplum">VMware Tanzu Greenplum</a> provides historic vehicle and track recording,searching, and data analysis</li>
            <li><a href="https://tanzu.vmware.com/spring-app-framework" target="_spring">Java Spring Boot</a> and <a href="https://steeltoe.io/" target="_steeltoe">C# .NET Core with Steeltoe</a> versions of each micro service are provided, showcasing the technology neutral platform</li>
            <li>A static ReactJS application (the one you're reading now!) is hosted by the VMware Tanzu Platform, downloaded to the user's web browser, and then communicates itself directly to the micro services it needs. This is licensed as an Apache 2.0 codebase and <a href="https://github.com/adamfowleruk/spring-adsb" target="_github">can be found on my GitHub repository</a></li>
          </ul>
        </Tab>
        <Tab eventKey='about-links' title="Links">
          <p>Links for the hosted version of this application:-</p>
          <ul>
            <li><a href="https://aircraft.shared.12factor.xyz" target="_hosted">Live Situation web application</a> hosted on AWS with the VMware Tanzu Platform</li>
            <li><a href="https://adsb.shared.12factor.xyz/data/aircraft.json" target="_liveview">RESTful data microservice live view JSON</a> hosted on the VMware Tanzu Platform on AWS</li>
          </ul>
        </Tab>
      </Tabs>
    )
  }
}
export default About;
