import React from 'react';

import API from '../api';

export default class Root extends React.Component {
    state = {
        stats: null,
    }

    componentDidMount() {
        API.get('/stats')
            .then((res) => {
                const stats = res.data;
                this.setState({stats})
            })
    }

    render() {
        return this.state.stats !== null ? (
            <div>
                <h3>Mongo {this.state.stats.mongo.ip}: {this.state.stats.mongo.visits}</h3>
                <h3>Redis {this.state.stats.redis.ip}: {this.state.stats.redis.visits}</h3>
            </div>
        ) : null;
    }
}